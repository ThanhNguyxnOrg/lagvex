//go:build linux

package client

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// RouteManager on Linux manages routes using `ip route`.
type RouteManager struct {
	mu             sync.Mutex
	tunName        string
	installedCIDRs []string
	pinnedRelay    string
	pinnedDev      string
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
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("running %v: %w (%s)", cmd, err, strings.TrimSpace(string(out)))
		}
	}

	log.Printf("[Route-Linux] Configured %s with IP %s (MTU %d)", r.tunName, cidr, mtu)
	return nil
}

func (r *RouteManager) PinRelayRoute(relayIP netip.Addr) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Detect default gateway: `ip route show default`
	cmd := exec.Command("ip", "route", "show", "default")
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
	_ = exec.Command("ip", "route", "del", prefix).Run()

	cmdAdd := exec.Command("ip", "route", "add", prefix, "via", gw, "dev", dev)
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

		_ = exec.Command("ip", "route", "del", cidr).Run()
		cmd := exec.Command("ip", "route", "add", cidr, "dev", r.tunName)
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
		_ = exec.Command("ip", "route", "del", cidr).Run()
	}
	r.installedCIDRs = r.installedCIDRs[:0]
}

func (r *RouteManager) RemoveAll() {
	r.RemoveGameRoutes()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.pinnedRelay != "" {
		_ = exec.Command("ip", "route", "del", r.pinnedRelay).Run()
		r.pinnedRelay = ""
	}
}

func (r *RouteManager) ActiveRouteCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := len(r.installedCIDRs)
	if r.pinnedRelay != "" {
		count++
	}
	return count
}
