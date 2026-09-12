//go:build windows

package client

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// RouteManager safely manages Windows network routes with store=active (RAM-only).
type RouteManager struct {
	mu sync.Mutex

	tunIfIndex       uint32
	installedGameCIDRs []string
	pinnedRelayRoute string
	pinnedPhysIndex  uint32
}

// NewRouteManager initializes a new RouteManager.
func NewRouteManager() *RouteManager {
	return &RouteManager{
		installedGameCIDRs: make([]string, 0),
	}
}

// ConfigureAdapter sets the static IPv4 address, netmask, MTU and disables DAD.
func (r *RouteManager) ConfigureAdapter(tunIfIndex uint32, innerIP netip.Addr, prefixLen, mtu int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tunIfIndex = tunIfIndex
	mask := prefixLengthToSubnetMask(prefixLen)

	// 1. Assign IP address
	cmdAddr := fmt.Sprintf("interface ipv4 set address name=%d source=static address=%s mask=%s store=active",
		tunIfIndex, innerIP.String(), mask)
	if err := runNetsh(cmdAddr); err != nil {
		return fmt.Errorf("set adapter address: %w", err)
	}

	// 2. Set MTU
	cmdMTU := fmt.Sprintf("interface ipv4 set subinterface interface=%d mtu=%d store=active",
		tunIfIndex, mtu)
	_ = runNetsh(cmdMTU)

	// 3. Disable Duplicate Address Detection (DAD) transmits
	cmdDAD := fmt.Sprintf("interface ipv4 set interface interface=%d dadtransmits=0 store=active",
		tunIfIndex)
	_ = runNetsh(cmdDAD)

	return nil
}

// PinRelayRoute pins a /32 route for the VPS relay through the physical network gateway.
// This prevents the critical routing loop where relay packets could get diverted into the tunnel.
func (r *RouteManager) PinRelayRoute(relayIP netip.Addr) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	physIndex, gateway, err := getDefaultGatewayRoute()
	if err != nil {
		return fmt.Errorf("detect default gateway: %w", err)
	}

	prefix := fmt.Sprintf("%s/32", relayIP.String())

	// Remove any previous pin if IP changed
	if r.pinnedRelayRoute != "" && r.pinnedRelayRoute != prefix {
		delCmd := fmt.Sprintf("interface ipv4 delete route prefix=%s interface=%d store=active",
			r.pinnedRelayRoute, r.pinnedPhysIndex)
		_ = runNetsh(delCmd)
		r.pinnedRelayRoute = ""
	}

	// Clear leftover before adding to avoid duplicate conflict
	delCmd := fmt.Sprintf("interface ipv4 delete route prefix=%s interface=%d store=active",
		prefix, physIndex)
	_ = runNetsh(delCmd)

	addCmd := fmt.Sprintf("interface ipv4 add route prefix=%s interface=%d nexthop=%s metric=1 store=active",
		prefix, physIndex, gateway.String())
	if err := runNetsh(addCmd); err != nil {
		return fmt.Errorf("pin relay route: %w", err)
	}

	r.pinnedRelayRoute = prefix
	r.pinnedPhysIndex = physIndex
	log.Printf("[Route] Pinned relay route %s via phys-interface %d (gw %s)", prefix, physIndex, gateway)
	return nil
}

// InstallGameRoutes installs on-link routes for the specified game CIDRs into the WinTun adapter.
func (r *RouteManager) InstallGameRoutes(cidrs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunIfIndex == 0 {
		log.Printf("[Route] Error: WinTun interface index is not set")
		return
	}

	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		// Check if already installed
		already := false
		for _, existing := range r.installedGameCIDRs {
			if existing == cidr {
				already = true
				break
			}
		}
		if already {
			continue
		}

		// Delete possible duplicate first
		_ = runNetsh(fmt.Sprintf("interface ipv4 delete route prefix=%s interface=%d store=active",
			cidr, r.tunIfIndex))

		// Add on-link route (no nexthop) with metric 1
		cmd := fmt.Sprintf("interface ipv4 add route prefix=%s interface=%d metric=1 store=active",
			cidr, r.tunIfIndex)
		if err := runNetsh(cmd); err == nil {
			r.installedGameCIDRs = append(r.installedGameCIDRs, cidr)
		} else {
			log.Printf("[Route] Failed to add route %s: %v", cidr, err)
		}
	}

	log.Printf("[Route] Installed %d game routes into WinTun (Interface %d)", len(r.installedGameCIDRs), r.tunIfIndex)
}

// RemoveGameRoutes removes all installed game CIDR routes.
func (r *RouteManager) RemoveGameRoutes() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunIfIndex == 0 || len(r.installedGameCIDRs) == 0 {
		return
	}

	for _, cidr := range r.installedGameCIDRs {
		cmd := fmt.Sprintf("interface ipv4 delete route prefix=%s interface=%d store=active",
			cidr, r.tunIfIndex)
		_ = runNetsh(cmd)
	}

	log.Printf("[Route] Removed %d game routes from WinTun", len(r.installedGameCIDRs))
	r.installedGameCIDRs = r.installedGameCIDRs[:0]
}

// RemoveAll cleans all game routes and unpins the relay route.
func (r *RouteManager) RemoveAll() {
	r.RemoveGameRoutes()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.pinnedRelayRoute != "" && r.pinnedPhysIndex != 0 {
		cmd := fmt.Sprintf("interface ipv4 delete route prefix=%s interface=%d store=active",
			r.pinnedRelayRoute, r.pinnedPhysIndex)
		_ = runNetsh(cmd)
		r.pinnedRelayRoute = ""
		r.pinnedPhysIndex = 0
		log.Printf("[Route] Unpinned relay route.")
	}
}

// ActiveRouteCount returns the number of active routes currently installed.
func (r *RouteManager) ActiveRouteCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := len(r.installedGameCIDRs)
	if r.pinnedRelayRoute != "" {
		count++
	}
	return count
}

func runNetsh(args string) error {
	cmd := exec.Command("netsh", strings.Fields(args)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func prefixLengthToSubnetMask(prefixLen int) string {
	mask := ^uint32(0) << (32 - prefixLen)
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(mask>>24), byte(mask>>16), byte(mask>>8), byte(mask))
}

// getDefaultGatewayRoute parses `route print 0.0.0.0` to find the default physical route.
func getDefaultGatewayRoute() (uint32, netip.Addr, error) {
	cmd := exec.Command("route", "print", "0.0.0.0")
	out, err := cmd.Output()
	if err != nil {
		return 0, netip.Addr{}, fmt.Errorf("route print: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	inActiveRoutes := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "Active Routes:") {
			inActiveRoutes = true
			continue
		}
		if inActiveRoutes {
			fields := strings.Fields(line)
			// e.g.: 0.0.0.0  0.0.0.0  192.168.1.1  192.168.1.50  25
			if len(fields) >= 5 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
				gwIP, err := netip.ParseAddr(fields[2])
				if err != nil {
					continue
				}
				nicIP, err := netip.ParseAddr(fields[3])
				if err != nil {
					continue
				}

				// Find interface index from IP
				ifIndex, err := findIfIndexByIP(nicIP)
				if err == nil {
					return ifIndex, gwIP, nil
				}
			}
		}
	}

	return 0, netip.Addr{}, fmt.Errorf("no active default gateway found in routing table")
}

func findIfIndexByIP(ip netip.Addr) (uint32, error) {
	cmd := exec.Command("netsh", "interface", "ipv4", "show", "addresses")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	currentIfIndex := uint32(0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Configuration for interface") {
			// Extract interface index if available
			idxParts := strings.Split(line, "\"")
			if len(idxParts) >= 2 {
				// Interface name or index
				currentIfIndex = getIfIndexByName(idxParts[1])
			}
		}
		if strings.Contains(line, "IP Address:") || strings.Contains(line, "IP:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 && parts[len(parts)-1] == ip.String() {
				if currentIfIndex != 0 {
					return currentIfIndex, nil
				}
			}
		}
	}

	// Fallback to searching netsh interface ipv4 show interfaces
	return getFirstPhysicalIfIndex()
}

func getIfIndexByName(name string) uint32 {
	cmd := exec.Command("netsh", "interface", "ipv4", "show", "interfaces")
	out, _ := cmd.Output()
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) >= 5 && strings.Contains(line, name) {
			if idx, err := strconv.ParseUint(fields[0], 10, 32); err == nil {
				return uint32(idx)
			}
		}
	}
	return 0
}

func getFirstPhysicalIfIndex() (uint32, error) {
	cmd := exec.Command("netsh", "interface", "ipv4", "show", "interfaces")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[2] == "connected" {
			if idx, err := strconv.ParseUint(fields[0], 10, 32); err == nil && idx > 1 {
				return uint32(idx), nil
			}
		}
	}
	return 0, fmt.Errorf("no connected physical interface found")
}
