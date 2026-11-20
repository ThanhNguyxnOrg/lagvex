//go:build darwin

package client

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"strings"
	"sync"
)

// RouteManager on macOS manages kernel routes via `route` and `ifconfig`.
type RouteManager struct {
	mu                  sync.Mutex
	tunName             string
	installedCIDRs      []string
	installedPeers      []string
	fullTunnelInstalled bool
	pinnedRelay         string
}

func NewRouteManager() *RouteManager {
	return &RouteManager{
		installedCIDRs: make([]string, 0),
	}
}

func (r *RouteManager) ConfigureAdapter(tunIfIndex uint32, innerIP netip.Addr, prefixLen, mtu int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find interface name from index
	iface, err := net.InterfaceByIndex(int(tunIfIndex))
	if err != nil {
		return fmt.Errorf("interface by index %d: %w", tunIfIndex, err)
	}
	r.tunName = iface.Name

	b := innerIP.As4()
	gwIP := netip.AddrFrom4([4]byte{b[0], b[1], b[2], 1})
	// ifconfig utunX <innerIP> <gwIP> mtu <mtu> up
	cmd := execHidden("ifconfig", r.tunName, innerIP.String(), gwIP.String(), "mtu", fmt.Sprintf("%d", mtu), "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig %s: %w (%s)", r.tunName, err, strings.TrimSpace(string(out)))
	}

	log.Printf("[Route-macOS] Configured %s with IP %s -> %s (MTU %d)", r.tunName, innerIP, gwIP, mtu)
	return nil
}

func (r *RouteManager) PinRelayRoute(relayIP netip.Addr) error {
	if relayIP.IsLoopback() {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Find default gateway via `netstat -nr -f inet` or `route get default`
	cmd := execHidden("route", "-n", "get", "default")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("get default route: %w", err)
	}

	var gateway string
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "gateway:" {
			gateway = fields[1]
			break
		}
	}
	if gateway == "" {
		return fmt.Errorf("default gateway not found")
	}

	// Remove any previous pin if IP changed
	if r.pinnedRelay != "" && r.pinnedRelay != relayIP.String() {
		_ = execHidden("route", "delete", "-host", r.pinnedRelay).Run()
		r.pinnedRelay = ""
	}

	// Delete existing pin before adding to avoid duplicate conflict
	_ = execHidden("route", "delete", "-host", relayIP.String()).Run()

	// route add -host <relayIP> <gateway>
	cmdAdd := execHidden("route", "add", "-host", relayIP.String(), gateway)
	if out, err := cmdAdd.CombinedOutput(); err != nil {
		return fmt.Errorf("pin relay route: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	r.pinnedRelay = relayIP.String()
	log.Printf("[Route-macOS] Pinned relay route %s via default gateway %s", relayIP, gateway)
	return nil
}

func (r *RouteManager) InstallGameRoutes(cidrs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunName == "" {
		return
	}

	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		_ = execHidden("route", "delete", "-net", cidr).Run()
		cmd := execHidden("route", "add", "-net", cidr, "-interface", r.tunName)
		if err := cmd.Run(); err == nil {
			r.installedCIDRs = append(r.installedCIDRs, cidr)
		}
	}

	log.Printf("[Route-macOS] Installed %d game routes onto %s", len(r.installedCIDRs), r.tunName)
}

func (r *RouteManager) RemoveGameRoutes() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cidr := range r.installedCIDRs {
		_ = execHidden("route", "delete", "-net", cidr).Run()
	}
	r.installedCIDRs = r.installedCIDRs[:0]
}

// AddPeerRoutes installs host routes for individual observed peer IPs and
// returns the addresses that were newly installed.
func (r *RouteManager) AddPeerRoutes(ips []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunName == "" {
		return nil
	}

	added := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		already := false
		for _, existing := range r.installedPeers {
			if existing == ip {
				already = true
				break
			}
		}
		if already {
			continue
		}
		_ = execHidden("route", "delete", "-host", ip).Run()
		if err := execHidden("route", "add", "-host", ip, "-interface", r.tunName).Run(); err == nil {
			r.installedPeers = append(r.installedPeers, ip)
			added = append(added, ip)
		}
	}
	return added
}

// RemovePeerRoutes removes previously installed peer host routes.
func (r *RouteManager) RemovePeerRoutes(ips []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(ips) == 0 {
		return
	}
	wanted := make(map[string]bool, len(ips))
	for _, ip := range ips {
		wanted[strings.TrimSpace(ip)] = true
	}
	kept := r.installedPeers[:0]
	for _, ip := range r.installedPeers {
		if wanted[ip] {
			_ = execHidden("route", "delete", "-host", ip).Run()
			continue
		}
		kept = append(kept, ip)
	}
	r.installedPeers = kept
}

func (r *RouteManager) removeAllPeersLocked() {
	for _, ip := range r.installedPeers {
		_ = execHidden("route", "delete", "-host", ip).Run()
	}
	r.installedPeers = r.installedPeers[:0]
}

// InstallFullTunnel routes all IPv4 traffic through the tunnel using the two
// half-default prefixes. The relay host route pin is untouched, so the tunnel
// itself never loops. Callers must treat this as an explicit user choice.
func (r *RouteManager) InstallFullTunnel() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunName == "" {
		return
	}
	for _, prefix := range fullTunnelPrefixes {
		_ = execHidden("route", "delete", "-net", prefix).Run()
		_ = execHidden("route", "add", "-net", prefix, "-interface", r.tunName).Run()
	}
	r.fullTunnelInstalled = true
}

// RemoveFullTunnel removes the half-default routes installed by InstallFullTunnel.
func (r *RouteManager) RemoveFullTunnel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeFullTunnelLocked()
}

func (r *RouteManager) removeFullTunnelLocked() {
	if !r.fullTunnelInstalled {
		return
	}
	for _, prefix := range fullTunnelPrefixes {
		_ = execHidden("route", "delete", "-net", prefix).Run()
	}
	r.fullTunnelInstalled = false
}

func (r *RouteManager) RemoveAll() {
	r.RemoveGameRoutes()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.removeAllPeersLocked()
	r.removeFullTunnelLocked()

	if r.pinnedRelay != "" {
		_ = execHidden("route", "delete", "-host", r.pinnedRelay).Run()
		r.pinnedRelay = ""
	}
}

func (r *RouteManager) ActiveRouteCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := len(r.installedCIDRs) + len(r.installedPeers)
	if r.pinnedRelay != "" {
		count++
	}
	return count
}
