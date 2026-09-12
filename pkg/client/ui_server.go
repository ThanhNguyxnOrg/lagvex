package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lagvex/lagvex/pkg/profiles"
	"github.com/lagvex/lagvex/pkg/protocol"
)

// UIServer serves the modern Gaming Booster Dashboard and REST API.
type UIServer struct {
	engine     *Engine
	profileMgr *profiles.Manager
	webDir     string
	server     *http.Server
}

// NewUIServer creates a new dashboard UI server.
func NewUIServer(engine *Engine, pm *profiles.Manager, webDir string) *UIServer {
	if webDir == "" {
		exe, _ := os.Executable()
		candidates := []string{
			filepath.Join(filepath.Dir(exe), "web"),
			`D:\Code\Lagvex\web`,
			"web",
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				webDir = c
				break
			}
		}
	}

	return &UIServer{
		engine:     engine,
		profileMgr: pm,
		webDir:     webDir,
	}
}

// Start listens on the specified address (e.g. "127.0.0.1:18888").
func (u *UIServer) Start(addr string) error {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/status", u.handleStatus)
	mux.HandleFunc("/api/games", u.handleGames)
	mux.HandleFunc("/api/relays", u.handleRelays)
	mux.HandleFunc("/api/connect", u.handleConnect)
	mux.HandleFunc("/api/disconnect", u.handleDisconnect)
	mux.HandleFunc("/api/test-relay", u.handleTestRelay)
	mux.HandleFunc("/api/add-game", u.handleAddGame)

	// Static Web Assets
	if u.webDir != "" {
		fs := http.FileServer(http.Dir(u.webDir))
		mux.Handle("/", fs)
	}

	u.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("[Dashboard] Web Dashboard active at http://%s", addr)
	return u.server.ListenAndServe()
}

func (u *UIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := u.engine.Stats()
	_ = json.NewEncoder(w).Encode(stats)
}

func (u *UIServer) handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	_ = json.NewEncoder(w).Encode(catalog.Games)
}

func (u *UIServer) handleRelays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	_ = json.NewEncoder(w).Encode(catalog.Relays)
}

type connectReq struct {
	RelayEndpoint string `json:"relayEndpoint"`
	PSK           string `json:"psk"`
	GameID        string `json:"gameId"`
	RegionID      string `json:"regionId"`
	ForceNow      bool   `json:"forceNow"`
}

func (u *UIServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req connectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.RelayEndpoint == "" {
		http.Error(w, "relayEndpoint is required", http.StatusBadRequest)
		return
	}

	go func() {
		if err := u.engine.Connect(req.RelayEndpoint, []byte(req.PSK), req.GameID, req.RegionID, req.ForceNow); err != nil {
			log.Printf("[Dashboard] Connect error: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "connecting"})
}

func (u *UIServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		_ = u.engine.Disconnect()
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "disconnecting"})
}

func (u *UIServer) handleTestRelay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
		PSK      string `json:"psk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Quick UDP latency ping test
	latencyMs, err := pingRelayTest(req.Endpoint, []byte(req.PSK))
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "latencyMs": latencyMs})
}

func (u *UIServer) handleAddGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var game profiles.GameDefinition
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if game.ID == "" || game.Name == "" {
		http.Error(w, "id and name are required", http.StatusBadRequest)
		return
	}

	if err := u.profileMgr.AddGame(game); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "game": game})
}

func pingRelayTest(endpoint string, psk []byte) (int64, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", endpoint)
	if err != nil {
		return 0, err
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	nonce, _ := protocol.RandomUint64()
	clientID, _ := protocol.RandomUint64()
	req := protocol.HandshakeRequest{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		ClientID:  clientID,
	}
	reqBuf := protocol.EncodeHandshakeRequest(psk, req)

	start := time.Now()
	_, _ = conn.WriteToUDP(reqBuf, udpAddr)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	recv := make([]byte, 128)
	n, _, err := conn.ReadFromUDP(recv)
	if err != nil {
		return 0, fmt.Errorf("timeout reaching relay")
	}

	elapsed := time.Since(start).Milliseconds()
	resp, err := protocol.DecodeHandshakeResponse(psk, recv[:n], nonce)
	if err != nil {
		return 0, fmt.Errorf("invalid handshake response (bad PSK?)")
	}
	if resp.Status != protocol.StatusOK {
		return 0, fmt.Errorf("relay returned status %d", resp.Status)
	}

	return elapsed, nil
}
