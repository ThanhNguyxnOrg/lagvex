package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// diagnosticsSnapshotVersion is the only snapshot schema this build reads.
const diagnosticsSnapshotVersion = 1

const diagnosticsMaxBodyBytes = 1 << 20 // 1 MiB

// diagnosticsKnownTweaks limits imported tweak keys to flags this client knows.
var diagnosticsKnownTweaks = map[string]bool{
	"tcpNoDelay":    true,
	"disableNagle":  true,
	"mmcssPriority": true,
	"mtuClamping":   true,
	"smartL4Filter": true,
}

type diagnosticsSystem struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Hostname   string `json:"hostname,omitempty"`
	Elevated   bool   `json:"elevated"`
	DriverMode string `json:"driverMode,omitempty"`
}

type diagnosticsRelay struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	Continent string `json:"continent,omitempty"`
	Location  string `json:"location,omitempty"`
}

type diagnosticsConfig struct {
	Nickname          string                        `json:"nickname"`
	LastSelectedGame  string                        `json:"lastSelectedGame,omitempty"`
	LastSelectedRelay string                        `json:"lastSelectedRelay,omitempty"`
	CustomGames       []profiles.GameDefinition     `json:"customGames"`
	CustomRelays      []diagnosticsRelay            `json:"customRelays"`
	Tweaks            map[string]bool               `json:"tweaks"`
	GameServers       map[string][]GameServerRecord `json:"gameServers,omitempty"`
}

type diagnosticsSnapshot struct {
	SnapshotVersion int                  `json:"snapshotVersion"`
	App             string               `json:"app"`
	GeneratedAt     time.Time            `json:"generatedAt"`
	PSKsIncluded    bool                 `json:"psksIncluded"`
	System          diagnosticsSystem    `json:"system"`
	Tunnel          TunnelStats          `json:"tunnel"`
	RelayHistory    []ProbeHistoryReport `json:"relayHistory"`
	Config          diagnosticsConfig    `json:"config"`
}

// handleDiagnosticsExport writes a JSON snapshot of measured state plus the
// user's non-secret settings. Relay pre-shared keys are never included.
func (u *UIServer) handleDiagnosticsExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname, _ := os.Hostname()
	stats := TunnelStats{State: StateDisconnected}
	if u.engine != nil {
		stats = u.engine.Stats()
	}

	cfg := UserConfig{}
	if u.configStore != nil {
		cfg = u.configStore.Get()
	}

	relays := make([]diagnosticsRelay, 0, len(cfg.CustomRelays))
	for _, rel := range cfg.CustomRelays {
		relays = append(relays, diagnosticsRelay{
			ID:        rel.ID,
			Name:      rel.Name,
			Endpoint:  rel.Endpoint,
			Continent: rel.Continent,
			Location:  rel.Location,
		})
	}

	history := []ProbeHistoryReport{}
	if u.prober != nil {
		history = u.prober.HistorySnapshot()
	}

	snapshot := diagnosticsSnapshot{
		SnapshotVersion: diagnosticsSnapshotVersion,
		App:             "lagvex-client",
		GeneratedAt:     time.Now().UTC(),
		PSKsIncluded:    false,
		System: diagnosticsSystem{
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			Hostname:   hostname,
			Elevated:   isProcessElevated(),
			DriverMode: stats.DriverMode,
		},
		Tunnel:       stats,
		RelayHistory: history,
		Config: diagnosticsConfig{
			Nickname:          cfg.Nickname,
			LastSelectedGame:  cfg.LastSelectedGame,
			LastSelectedRelay: cfg.LastSelectedRelay,
			CustomGames:       cfg.CustomGames,
			CustomRelays:      relays,
			Tweaks:            cfg.Tweaks,
			GameServers:       cfg.GameServers,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=lagvex-diagnostics.json")
	_ = json.NewEncoder(w).Encode(snapshot)
}

type diagnosticsImportResult struct {
	Success  bool     `json:"success"`
	Applied  []string `json:"applied"`
	Warnings []string `json:"warnings,omitempty"`
}

// handleDiagnosticsImport merges a snapshot's non-secret settings back in.
// Relay entries are reported as skipped because snapshots never carry PSKs.
func (u *UIServer) handleDiagnosticsImport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if u.configStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "settings store is unavailable")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, diagnosticsMaxBodyBytes)
	var snap diagnosticsSnapshot
	if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid snapshot JSON: "+err.Error())
		return
	}
	if snap.SnapshotVersion != diagnosticsSnapshotVersion {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("unsupported snapshotVersion %d", snap.SnapshotVersion))
		return
	}

	result := diagnosticsImportResult{Success: true}

	if snap.Config.Nickname != "" {
		u.configStore.SetNickname(snap.Config.Nickname)
		result.Applied = append(result.Applied, "nickname")
	}

	if len(snap.Config.Tweaks) > 0 {
		known := map[string]bool{}
		for key, value := range snap.Config.Tweaks {
			if diagnosticsKnownTweaks[key] {
				known[key] = value
			} else {
				result.Warnings = append(result.Warnings, fmt.Sprintf("ignored unknown tweak %q", key))
			}
		}
		if len(known) > 0 {
			u.configStore.SetTweaks(known)
			result.Applied = append(result.Applied, fmt.Sprintf("%d tweak flag(s)", len(known)))
		}
	}

	gamesApplied := 0
	for _, game := range snap.Config.CustomGames {
		if strings.TrimSpace(game.ID) == "" || strings.TrimSpace(game.Name) == "" {
			result.Warnings = append(result.Warnings, "skipped a custom game without id or name")
			continue
		}
		hasCIDR := false
		for _, region := range game.Regions {
			if len(region.CIDRs) > 0 {
				hasCIDR = true
				break
			}
		}
		if !hasCIDR {
			result.Warnings = append(result.Warnings, fmt.Sprintf("skipped custom game %q: no region CIDRs", game.ID))
			continue
		}
		u.configStore.AddCustomGame(game)
		if u.profileMgr != nil {
			_ = u.profileMgr.AddGame(game)
		}
		gamesApplied++
	}
	if gamesApplied > 0 {
		result.Applied = append(result.Applied, fmt.Sprintf("%d custom game(s)", gamesApplied))
	}

	if len(snap.Config.CustomRelays) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"skipped %d relay entr(ies): snapshots never contain pre-shared keys; add them again with their PSK",
			len(snap.Config.CustomRelays)))
	}

	if snap.Config.LastSelectedGame != "" || snap.Config.LastSelectedRelay != "" {
		u.configStore.SetSelections(snap.Config.LastSelectedGame, snap.Config.LastSelectedRelay)
		result.Applied = append(result.Applied, "last selections")
	}

	// Game-server history carries only measured endpoints (positive avgMs was
	// required when each record was written), so merging is safe.
	serversMerged := 0
	for gameID, records := range snap.Config.GameServers {
		if strings.TrimSpace(gameID) == "" {
			continue
		}
		for _, rec := range records {
			if rec.Host == "" || rec.AvgMs <= 0 {
				continue
			}
			u.configStore.RecordGameServer(gameID, rec)
			serversMerged++
		}
	}
	if serversMerged > 0 {
		result.Applied = append(result.Applied, fmt.Sprintf("%d game-server record(s)", serversMerged))
	}

	_ = json.NewEncoder(w).Encode(result)
}
