package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"testing"
	"time"
)

func TestRelaySquadSignaling(t *testing.T) {
	cfg := Config{
		ListenAddr:  "127.0.0.1:44990",
		HTTPAddr:    "127.0.0.1:44991",
		TunName:     "lagvex-test",
		Subnet:      netip.MustParsePrefix("10.99.0.0/24"),
		PSK:         []byte("test-psk-12345"),
		MTU:         1400,
		IdleTimeout: 10 * time.Second,
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()

	baseURL := "http://127.0.0.1:44991"

	// 1. Health check with resilient retry (wait up to 2 seconds for server to bind)
	var resp *http.Response
	var lastErr error
	for i := 0; i < 40; i++ {
		resp, lastErr = http.Get(baseURL + "/health")
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr != nil {
		t.Fatalf("health check failed: %v", lastErr)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on health check, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Create Squad Room
	createBody, _ := json.Marshal(map[string]any{
		"code": "LGVX-RELAY",
		"name": "HostPlayer#1001",
		"game": "Valorant",
		"ping": 16,
		"isp":  "Fiber",
	})
	resp, err = http.Post(baseURL+"/squad/create", "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("squad create failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on squad create, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. Join Squad Room
	joinBody, _ := json.Marshal(map[string]any{
		"code": "LGVX-RELAY",
		"name": "Teammate#2002",
		"game": "Valorant",
		"ping": 22,
		"isp":  "Broadband",
	})
	resp, err = http.Post(baseURL+"/squad/join", "application/json", bytes.NewReader(joinBody))
	if err != nil {
		t.Fatalf("squad join failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on squad join, got %d", resp.StatusCode)
	}
	var joinResult struct {
		Members []RelaySquadMember `json:"members"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&joinResult)
	resp.Body.Close()

	if len(joinResult.Members) != 2 {
		t.Fatalf("expected 2 members in squad room, got %d", len(joinResult.Members))
	}

	// 4. Query Room
	resp, err = http.Get(baseURL + "/squad/room?code=LGVX-RELAY")
	if err != nil {
		t.Fatalf("query room failed: %v", err)
	}
	var roomResult struct {
		Members []RelaySquadMember `json:"members"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&roomResult)
	resp.Body.Close()

	if len(roomResult.Members) != 2 {
		t.Fatalf("expected 2 members from query room, got %d", len(roomResult.Members))
	}
}
