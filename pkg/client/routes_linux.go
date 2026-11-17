//go:build linux

package client

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
)

// RouteManager on Linux manages routes using `ip route`.
type RouteManager struct {
	mu                  sync.Mutex
	tunName             string
	installedCIDRs      []string
	installedPeers      []string
	fullTunnelInstalled bool
	pinnedRelay         string
	pinnedDev           string
}

func NewRouteManager() *RouteManager {
	return &RouteManager{
		installedCIDRs: make([]string, 0),
	}
}

func (r *RouteManager) ConfigureAdapter(tunIfIndex uint32, innerIP netip.Addr, prefixLen, mtu int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	iface, err := net.InterfaceByIndex(int(tunIfIndex))
	if err != nil {
		return fmt.Errorf("find iface %d: %w", tunIfIndex, err)
	}
	r.tunName = iface.Name

	cidr := fmt.Sprintf("%s/%d", innerIP.String(), prefixLen)
	commands := [][]string{
		{"ip", "addr", "replace", cidr, "dev", r.tunName},
		{"ip", "link", "set", "dev", r.tunName, "mtu", strconv.Itoa(mtu)},
		{"ip", "link", "set", "dev", r.tunName, "up"},
	}

	for _, cmd := range commands {
		if out, err := execHidden(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("running %v: %w (%s)", cmd, err, strings.TrimSpace(string(out)))
		}
	}

	log.Printf("[Route-Linux] Configured %s with IP %s (MTU %d)", r.tunName, cidr, mtu)
	return nil
}

func (r *RouteManager) PinRelayRoute(relayIP netip.Addr) error {
	if relayIP.IsLoopback() {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Detect default gateway: `ip route show default`
	cmd := execHidden("ip", "route", "show", "default")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("find default route: %w", err)
	}

	fields := strings.Fields(string(out))
	var gw, dev string
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "via" {
			gw = fields[i+1]
		}
		if fields[i] == "dev" {
			dev = fields[i+1]
		}
	}
	if gw == "" || dev == "" {
		return fmt.Errorf("default gateway or dev not found")
	}

	prefix := fmt.Sprintf("%s/32", relayIP.String())

	// Remove previous pin if IP changed
	if r.pinnedRelay != "" && r.pinnedRelay != prefix {
		_ = execHidden("ip", "route", "del", r.pinnedRelay).Run()
		r.pinnedRelay = ""
	}

	_ = execHidden("ip", "route", "del", prefix).Run()

	cmdAdd := execHidden("ip", "route", "add", prefix, "via", gw, "dev", dev)
	if out, err := cmdAdd.CombinedOutput(); err != nil {
		return fmt.Errorf("pin relay route: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	r.pinnedRelay = prefix
	r.pinnedDev = dev
	log.Printf("[Route-Linux] Pinned relay route %s via %s dev %s", prefix, gw, dev)
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

		_ = execHidden("ip", "route", "del", cidr).Run()
		cmd := execHidden("ip", "route", "add", cidr, "dev", r.tunName)
		if err := cmd.Run(); err == nil {
			r.installedCIDRs = append(r.installedCIDRs, cidr)
		}
	}

	log.Printf("[Route-Linux] Installed %d game routes on %s", len(r.installedCIDRs), r.tunName)
}

func (r *RouteManager) RemoveGameRoutes() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunName == "" {
		return
	}

	for _, cidr := range r.installedCIDRs {
		_ = execHidden("ip", "route", "del", cidr).Run()
	}
	r.installedCIDRs = r.installedCIDRs[:0]
}

// AddPeerRoutes installs /32 routes for individual observed peer IPs and
// returns the prefixes that were newly installed.
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
		prefix := ip + "/32"
		already := false
		for _, existing := range r.installedPeers {
			if existing == prefix {
				already = true
				break
			}
		}
		if already {
			continue
		}
		if err := execHidden("ip", "route", "replace", prefix, "dev", r.tunName).Run(); err == nil {
			r.installedPeers = append(r.installedPeers, prefix)
			added = append(added, prefix)
		}
	}
	return added
}

// RemovePeerRoutes removes previously installed peer /32 routes.
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
	for _, prefix := range r.installedPeers {
		if wanted[strings.TrimSuffix(prefix, "/32")] {
			_ = execHidden("ip", "route", "del", prefix).Run()
			continue
		}
		kept = append(kept, prefix)
	}
	r.installedPeers = kept
}

func (r *RouteManager) removeAllPeersLocked() {
	for _, prefix := range r.installedPeers {
		_ = execHidden("ip", "route", "del", prefix).Run()
	}
	r.installedPeers = r.installedPeers[:0]
}

// InstallFullTunnel routes all IPv4 traffic through the tunnel using the two
// half-default prefixes. The relay /32 pin is untouched, so the tunnel itself
// never loops. Callers must treat this as an explicit user choice.
func (r *RouteManager) InstallFullTunnel() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tunName == "" {
		return
	}
	for _, prefix := range fullTunnelPrefixes {
		_ = execHidden("ip", "route", "replace", prefix, "dev", r.tunName).Run()
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
		_ = execHidden("ip", "route", "del", prefix).Run()
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
		_ = execHidden("ip", "route", "del", r.pinnedRelay).Run()
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
