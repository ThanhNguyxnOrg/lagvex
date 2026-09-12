//go:build darwin

package client

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"strings"
	"sync"
)

// RouteManager on macOS manages kernel routes via `route` and `ifconfig`.
type RouteManager struct {
	mu           sync.Mutex
	tunName      string
	installedCIDRs []string
	pinnedRelay  string
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

	gwIP := innerIP.Next()
	// ifconfig utunX <innerIP> <gwIP> mtu <mtu> up
	cmd := exec.Command("ifconfig", r.tunName, innerIP.String(), gwIP.String(), "mtu", fmt.Sprintf("%d", mtu), "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ifconfig %s: %w (%s)", r.tunName, err, strings.TrimSpace(string(out)))
	}

	log.Printf("[Route-macOS] Configured %s with IP %s -> %s (MTU %d)", r.tunName, innerIP, gwIP, mtu)
	return nil
}

func (r *RouteManager) PinRelayRoute(relayIP netip.Addr) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find default gateway via `netstat -nr -f inet` or `route get default`
	cmd := exec.Command("route", "-n", "get", "default")
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

	// Delete existing pin
	_ = exec.Command("route", "delete", "-host", relayIP.String()).Run()

	// route add -host <relayIP> <gateway>
	cmdAdd := exec.Command("route", "add", "-host", relayIP.String(), gateway)
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

		_ = exec.Command("route", "delete", "-net", cidr).Run()
		cmd := exec.Command("route", "add", "-net", cidr, "-interface", r.tunName)
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
		_ = exec.Command("route", "delete", "-net", cidr).Run()
	}
	r.installedCIDRs = r.installedCIDRs[:0]
}

func (r *RouteManager) RemoveAll() {
	r.RemoveGameRoutes()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.pinnedRelay != "" {
		_ = exec.Command("route", "delete", "-host", r.pinnedRelay).Run()
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
