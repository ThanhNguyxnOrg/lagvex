package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// routeHop is one measured (or explicitly unmeasured) segment of the path
// between this PC and the game server.
type routeHop struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Target   string  `json:"target,omitempty"`
	RTTMs    float64 `json:"rttMs,omitempty"`
	RTTAvgMs float64 `json:"rttAvgMs,omitempty"`
	RTTStdMs float64 `json:"rttStdMs,omitempty"`
	LossPct  float64 `json:"lossPct"`
	Samples  int     `json:"samples,omitempty"`
	Measured bool    `json:"measured"`
	Reason   string  `json:"reason,omitempty"`
	Note     string  `json:"note,omitempty"`
}

// handleRouteDetail measures the path pieces Lagvex can actually observe:
// this PC -> local gateway, PC -> relay (direct probe and live tunnel echo),
// and a public cloud anchor in the game region. The relay -> game server
// segment is reported as unmeasured instead of being invented.
func (u *UIServer) handleRouteDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameID := strings.TrimSpace(r.URL.Query().Get("gameId"))
	regionID := strings.TrimSpace(r.URL.Query().Get("regionId"))
	relayEndpoint := strings.TrimSpace(r.URL.Query().Get("relayEndpoint"))

	regionName := ""
	if gameID != "" {
		if regionID == "" {
			writeJSONError(w, http.StatusBadRequest, "regionId is required when gameId is set")
			return
		}
		if u.profileMgr == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "game profiles are unavailable")
			return
		}
		game, ok := u.profileMgr.FindGameByID(gameID)
		if !ok {
			writeJSONError(w, http.StatusBadRequest, "unknown gameId")
			return
		}
		found := false
		for _, region := range game.Regions {
			if strings.EqualFold(region.ID, regionID) {
				regionName = region.Name
				found = true
				break
			}
		}
		if !found {
			writeJSONError(w, http.StatusBadRequest, "unknown regionId for this game")
			return
		}
	}

	var stats TunnelStats
	if u.engine != nil {
		stats = u.engine.Stats()
	}
	connected := stats.State == StateConnected
	if relayEndpoint == "" && connected {
		relayEndpoint = stats.RelayAddr
	}

	hops := make([]routeHop, 0, 5)
	hops = append(hops, routeHop{
		ID:       "pc",
		Label:    "This PC",
		Target:   "127.0.0.1",
		Measured: true,
		Note:     "traffic origin",
	})

	// Hop: local gateway. Measured with a TCP handshake because raw ICMP
	// needs extra privileges; port 80 has to answer for a value to exist.
	gatewayHop := routeHop{ID: "gateway", Label: "Local gateway"}
	if gateway := detectDefaultGateway(); gateway != "" {
		gatewayHop.Target = gateway
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		rtts, _ := probeTCPGateway(ctx, net.JoinHostPort(gateway, "80"), 2)
		cancel()
		if len(rtts) > 0 {
			gatewayHop.RTTMs = roundTo(calculateMedian(rtts), 1)
			gatewayHop.Samples = len(rtts)
			gatewayHop.Measured = true
			gatewayHop.Note = "TCP connect to port 80"
		} else {
			gatewayHop.Reason = "gateway did not answer a TCP connect on port 80"
		}
	} else {
		gatewayHop.Reason = "default gateway could not be detected"
	}
	hops = append(hops, gatewayHop)

	// Hop: relay, measured directly (pre-tunnel path) plus any retained
	// history for this endpoint.
	relayHop := routeHop{ID: "relay", Label: "Relay node", Target: relayEndpoint}
	if relayEndpoint == "" {
		relayHop.Reason = "no relay is selected or connected"
	} else {
		var catalogPSK string
		if u.profileMgr != nil {
			for _, relay := range u.profileMgr.Catalog().Relays {
				if strings.EqualFold(relay.Endpoint, relayEndpoint) {
					catalogPSK = relay.PSK
					break
				}
			}
		}
		if catalogPSK == "" {
			relayHop.Reason = "relay PSK is unavailable; add this relay with its PSK first"
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			result := u.prober.ProbeEndpoint(ctx, profiles.RelayEndpoint{
				ID:       "route-detail",
				Endpoint: relayEndpoint,
				PSK:      catalogPSK,
			}, 2)
			cancel()
			if result.Reachable {
				relayHop.RTTMs = result.RTTMedianMs
				relayHop.LossPct = result.PacketLoss
				relayHop.Samples = result.Samples
				relayHop.Measured = true
				relayHop.Note = "direct UDP probe, bypassing the tunnel"
			} else {
				relayHop.Reason = "direct probe failed: " + result.Error
			}
			if history, ok := u.prober.HistoryFor(relayEndpoint); ok {
				relayHop.RTTAvgMs = history.RTTAvgMs
				relayHop.RTTStdMs = history.RTTStdMs
			}
		}
	}
	hops = append(hops, relayHop)

	// Hop: the live tunnel itself (engine echo RTT through the encrypted
	// tunnel). Comparable with the direct relay probe above.
	tunnelHop := routeHop{ID: "tunnel", Label: "Tunnel (PC to relay, encrypted)"}
	if connected && stats.PingMs > 0 {
		tunnelHop.RTTMs = float64(stats.PingMs)
		tunnelHop.LossPct = stats.PacketLossPct
		tunnelHop.Measured = true
		tunnelHop.Note = "live engine echo"
	} else {
		tunnelHop.Reason = "no live tunnel RTT sample yet"
	}
	hops = append(hops, tunnelHop)

	// Hop: game server. The relay can measure its own RTT to an allowlisted
	// anchor (see "anchor" below), but it never probes the actual game server,
	// so this segment is not guessed.
	gameServerHop := routeHop{
		ID:     "game-server",
		Label:  "Game server",
		Target: regionName,
		Reason: "no server-side probe target; only the region anchor below is measured",
	}
	hops = append(hops, gameServerHop)

	overheadMs := 0.0
	if tunnelHop.Measured && relayHop.Measured && tunnelHop.RTTMs > 0 && relayHop.RTTMs > 0 {
		overheadMs = roundTo(tunnelHop.RTTMs-relayHop.RTTMs, 1)
	}

	type anchorStatus struct {
		ID                  string  `json:"id"`
		Label               string  `json:"label"`
		Target              string  `json:"target"`
		RTTMs               float64 `json:"rttMs,omitempty"`
		Measured            bool    `json:"measured"`
		Note                string  `json:"note,omitempty"`
		RelayRTTMs          float64 `json:"relayRttMs,omitempty"`
		RelayMeasured       bool    `json:"relayMeasured"`
		ViaRelayEstimateMs  float64 `json:"viaRelayEstimateMs,omitempty"`
		ImprovementMs       float64 `json:"improvementMs,omitempty"`
		ImprovementMeasured bool    `json:"improvementMeasured"`
	}
	anchor := anchorStatus{}
	if anchorMatch, ok := matchTelemetryAnchor(regionID, regionName); ok {
		anchor.ID = anchorMatch.id
		anchor.Label = anchorMatch.label
		anchor.Target = anchorTarget(anchorMatch)
		anchor.Note = "public cloud reference for this region, not the game server"

		if target := anchorTarget(anchorMatch); target != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			rtts, _ := probeTCPGateway(ctx, target, 2)
			cancel()
			if len(rtts) > 0 {
				anchor.RTTMs = roundTo(calculateMedian(rtts), 1)
				anchor.Measured = true
			}
		}

		// Ask the relay to measure its own side of the same anchor. The relay
		// only accepts allowlisted ids, and old relays simply do not answer.
		if connected && u.engine != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
			relayRTT, _, err := u.engine.ProbeAnchorViaRelay(ctx, anchorMatch.probeID)
			cancel()
			if err == nil && relayRTT > 0 {
				anchor.RelayRTTMs = roundTo(relayRTT, 1)
				anchor.RelayMeasured = true
			}
		}
	}

	// The via-relay estimate is the sum of two measured hops (PC -> relay plus
	// relay -> anchor). It only appears when both measurements exist, and the
	// improvement compares it with the direct anchor RTT.
	if anchor.RelayMeasured && tunnelHop.Measured && tunnelHop.RTTMs > 0 {
		anchor.ViaRelayEstimateMs = roundTo(tunnelHop.RTTMs+anchor.RelayRTTMs, 1)
		if anchor.Measured && anchor.RTTMs > 0 {
			anchor.ImprovementMs = roundTo(anchor.RTTMs-anchor.ViaRelayEstimateMs, 1)
			anchor.ImprovementMeasured = true
		}
	}

	notes := []string{
		"RTTs are measured with TCP handshakes or the tunnel echo, not ICMP, so they can differ from in-game ping.",
	}
	if !connected {
		notes = append(notes, "Connect to a relay to add the live tunnel measurement.")
	}
	if anchor.ImprovementMeasured {
		notes = append(notes, "Improvement compares the direct anchor RTT with tunnel RTT + relay-measured anchor RTT; it is a region estimate, not a game-server measurement.")
	}
	if connected && !anchor.RelayMeasured {
		notes = append(notes, "The relay did not answer the anchor probe (older relay build or rate limit), so the estimate is unavailable.")
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"measuredAt": time.Now().UTC().Format(time.RFC3339),
		"gameId":     gameID,
		"regionId":   regionID,
		"regionName": regionName,
		"connected":  connected,
		"hops":       hops,
		"overheadMs": overheadMs,
		"anchor":     anchor,
		"notes":      notes,
	})
}
