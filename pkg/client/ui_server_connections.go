package client

import (
	"encoding/json"
	"net/http"
	"time"
)

type gameConnectionStatus struct {
	GameID      string              `json:"gameId"`
	Name        string              `json:"name"`
	Running     bool                `json:"running"`
	Process     string              `json:"process,omitempty"`
	PID         int                 `json:"pid,omitempty"`
	Connections []ProcessConnection `json:"connections"`
}

// handleGameConnections lists the remote endpoints held by running game
// processes, read from the local OS socket tables only. It never attaches to,
// injects into, or reads memory of a game process.
func (u *UIServer) handleGameConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()

	var targets []string
	for _, game := range catalog.Games {
		targets = append(targets, game.ProcessNames...)
	}
	matches := findRunningProcesses(targets)

	games := make([]gameConnectionStatus, 0, len(catalog.Games))
	supported := true
	for _, game := range catalog.Games {
		status := gameConnectionStatus{
			GameID:      game.ID,
			Name:        game.Name,
			Connections: []ProcessConnection{},
		}
		for _, proc := range game.ProcessNames {
			pid, ok := matches[proc]
			if !ok {
				continue
			}
			status.Running = true
			status.Process = proc
			status.PID = pid
			conns, ok := scanProcessConnections(pid)
			if !ok {
				supported = false
			}
			status.Connections = mergeConnections(status.Connections, conns)
			break
		}
		games = append(games, status)
	}

	resp := map[string]any{
		"scannedAt": time.Now().UTC().Format(time.RFC3339),
		"supported": supported,
		"note":      connectionScanNote,
		"games":     games,
	}
	if !supported {
		resp["reason"] = "socket enumeration is unavailable on this system"
	}
	_ = json.NewEncoder(w).Encode(resp)
}
