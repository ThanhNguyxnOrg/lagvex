package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

func TestRegisterFreeWARP(t *testing.T) {
	peerPub := base64.StdEncoding.EncodeToString([]byte("1234567890abcdef1234567890abcdef"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var req warpRegRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if _, err := profiles.DecodeWireGuardKeyHex(req.Key); err != nil {
			t.Errorf("invalid generated public key: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "test-warp-device",
			"config": map[string]any{
				"peers": []map[string]any{
					{
						"public_key": peerPub,
						"endpoint": map[string]any{
							"v4":   "162.159.192.1:0",
							"host": "engage.cloudflareclient.com:2408",
						},
					},
				},
				"interface": map[string]any{
					"addresses": map[string]any{
						"v4": "172.16.0.2",
					},
				},
			},
		})
	}))
	defer srv.Close()

	old := warpRegEndpoint
	warpRegEndpoint = srv.URL
	defer func() { warpRegEndpoint = old }()

	relay, err := RegisterFreeWARP(context.Background())
	if err != nil {
		t.Fatalf("RegisterFreeWARP failed: %v", err)
	}
	if !relay.IsWireGuard() {
		t.Fatalf("expected wireguard transport")
	}
	if relay.Endpoint != "162.159.192.1:2408" {
		t.Fatalf("endpoint = %q, want 162.159.192.1:2408", relay.Endpoint)
	}
	if relay.WireGuard.Address != "172.16.0.2/32" {
		t.Fatalf("address = %q, want 172.16.0.2/32", relay.WireGuard.Address)
	}

	uapi, err := buildWireGuardUAPI(relay.WireGuard, netip.MustParseAddrPort(relay.Endpoint))
	if err != nil {
		t.Fatalf("buildWireGuardUAPI failed: %v", err)
	}
	if len(uapi) == 0 {
		t.Fatalf("expected non-empty UAPI config")
	}
}
