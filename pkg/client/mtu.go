package client

import (
	"context"
	"net/netip"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

// MTUDiscoveryResult reports the measured non-fragmenting path MTU and derived
// safe tunnel MTUs for both native Lagvex and WireGuard transports.
type MTUDiscoveryResult struct {
	Host           string `json:"host"`
	PathMTU        int    `json:"pathMtu"`
	RecommendedMTU int    `json:"recommendedMtu"` // Lagvex AEAD (IPv4 20 + UDP 8 + AEAD 33 = 61B overhead)
	WireGuardMTU   int    `json:"wireguardMtu"`   // WireGuard (IPv4 20 + UDP 8 + WG 32 + safety = 80B overhead)
	Probes         int    `json:"probes"`
	Measured       bool   `json:"measured"`
	MeasuredAt     string `json:"measuredAt"`
	DurationMs     int64  `json:"durationMs"`
	Note           string `json:"note,omitempty"`
}

// dfPingProbe tests whether an IPv4 ICMP packet with the Don't Fragment bit set
// and the given ICMP payload size reaches ip without fragmentation.
// Overridable in unit tests so tests never send real ICMP traffic.
var dfPingProbe = func(ctx context.Context, ip netip.Addr, payloadBytes int) bool {
	var cmd *exec.Cmd
	sizeStr := strconv.Itoa(payloadBytes)
	switch runtime.GOOS {
	case "windows":
		cmd = execHiddenContext(ctx, "ping", "-n", "1", "-f", "-l", sizeStr, "-w", "900", ip.String())
	case "darwin":
		cmd = execHiddenContext(ctx, "ping", "-c", "1", "-D", "-s", sizeStr, "-W", "900", ip.String())
	default:
		if path, err := exec.LookPath("ping"); err == nil {
			cmd = execHiddenContext(ctx, path, "-c", "1", "-M", "do", "-s", sizeStr, "-W", "1", ip.String())
		} else {
			return false
		}
	}
	return cmd.Run() == nil
}

// discoverPathMTU runs a binary search over IPv4 DF-bit ICMP payload sizes
// [1252..1472] (corresponding to IPv4 path MTU [1280..1500]) to find the
// highest non-fragmenting MTU on the current ISP link (e.g. 1500 Ethernet,
// 1492 PPPoE, 1420 cellular).
func discoverPathMTU(ctx context.Context, ip netip.Addr) MTUDiscoveryResult {
	start := time.Now()
	res := MTUDiscoveryResult{
		Host:       ip.String(),
		MeasuredAt: time.Now().UTC().Format(time.RFC3339),
	}

	const (
		ipICMPHeader = 28   // 20-byte IPv4 header + 8-byte ICMP echo header
		minPayload   = 1252 // 1280 MTU
		maxPayload   = 1472 // 1500 MTU
	)

	// Fast check: verify the minimum payload succeeds first so a blocked-ICMP
	// network fails cleanly instead of reporting a bogus MTU.
	res.Probes++
	if !dfPingProbe(ctx, ip, minPayload) {
		res.DurationMs = time.Since(start).Milliseconds()
		res.Note = "ICMP DF probe did not answer (firewall or ICMP blocked on this network)"
		return res
	}

	// Fast path: standard Ethernet 1500 MTU succeeds on probe 2.
	res.Probes++
	if dfPingProbe(ctx, ip, maxPayload) {
		return finalizeMTUResult(res, maxPayload+ipICMPHeader, start)
	}

	lo, hi := minPayload, maxPayload-1
	best := minPayload
	for lo <= hi {
		if ctx.Err() != nil {
			break
		}
		mid := (lo + hi) / 2
		res.Probes++
		if dfPingProbe(ctx, ip, mid) {
			best = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return finalizeMTUResult(res, best+ipICMPHeader, start)
}

func finalizeMTUResult(res MTUDiscoveryResult, pathMTU int, start time.Time) MTUDiscoveryResult {
	res.Measured = true
	res.PathMTU = pathMTU
	// Lagvex AEAD overhead: 20B IPv4 + 8B UDP + 33B AEAD (+ 8B optional dual trailer = 69B margin)
	lagvexMTU := pathMTU - 72
	if lagvexMTU > 1400 {
		lagvexMTU = 1400
	}
	if lagvexMTU < 1200 {
		lagvexMTU = 1200
	}
	wgMTU := pathMTU - 80
	if wgMTU < 1200 {
		wgMTU = 1200
	}
	res.RecommendedMTU = lagvexMTU
	res.WireGuardMTU = wgMTU
	res.DurationMs = time.Since(start).Milliseconds()
	switch {
	case pathMTU >= 1500:
		res.Note = "Full 1500B Ethernet path MTU detected (no PPPoE bottleneck)"
	case pathMTU >= 1492:
		res.Note = "1492B PPPoE path MTU detected — tunnel MTU tuned to avoid fragmentation"
	default:
		res.Note = "Reduced path MTU detected (VPN/cellular/encapsulated ISP link)"
	}
	return res
}
