package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"golang.org/x/crypto/curve25519"
)

// defaultWARPRegEndpoint is Cloudflare's public WARP client registration URL.
// Overridable in tests so unit tests never reach the live internet.
var warpRegEndpoint = "https://api.cloudflareclient.com/v0a2158/reg"

// defaultWARPFallbackEndpoint is Cloudflare's documented Anycast WireGuard edge.
const defaultWARPFallbackEndpoint = "162.159.192.1:2408"

// generateWireGuardKeypair creates a fresh Curve25519 keypair encoded in base64.
func generateWireGuardKeypair() (privB64, pubB64 string, err error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return "", "", fmt.Errorf("read entropy: %w", err)
	}
	// Standard X25519 clamping per RFC 7748.
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pub, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return "", "", fmt.Errorf("derive public key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(priv[:]), base64.StdEncoding.EncodeToString(pub), nil
}

type warpRegRequest struct {
	Key       string `json:"key"`
	InstallID string `json:"install_id"`
	FCMToken  string `json:"fcm_token"`
	TOS       string `json:"tos"`
	Type      string `json:"type"`
	Model     string `json:"model"`
	Locale    string `json:"locale"`
}

type warpRegResponse struct {
	ID     string `json:"id"`
	Config struct {
		Peers []struct {
			PublicKey string `json:"public_key"`
			Endpoint  struct {
				V4   string `json:"v4"`
				Host string `json:"host"`
			} `json:"endpoint"`
		} `json:"peers"`
		Interface struct {
			Addresses struct {
				V4 string `json:"v4"`
			} `json:"addresses"`
		} `json:"interface"`
	} `json:"config"`
}

// RegisterFreeWARP creates a fresh free Cloudflare WARP WireGuard profile that
// can be used as a zero-cost split-tunnel backbone inside Lagvex.
func RegisterFreeWARP(ctx context.Context) (profiles.RelayEndpoint, error) {
	privB64, pubB64, err := generateWireGuardKeypair()
	if err != nil {
		return profiles.RelayEndpoint{}, err
	}

	payload := warpRegRequest{
		Key:       pubB64,
		InstallID: "",
		FCMToken:  "",
		TOS:       time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Type:      "PC",
		Model:     "Lagvex",
		Locale:    "en_US",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return profiles.RelayEndpoint{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, warpRegEndpoint, bytes.NewReader(body))
	if err != nil {
		return profiles.RelayEndpoint{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User-Agent", "okhttp/3.12.1")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return profiles.RelayEndpoint{}, fmt.Errorf("contact Cloudflare WARP registration API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return profiles.RelayEndpoint{}, fmt.Errorf("read WARP registration response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return profiles.RelayEndpoint{}, fmt.Errorf("Cloudflare WARP registration returned HTTP %d", resp.StatusCode)
	}

	var parsed warpRegResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return profiles.RelayEndpoint{}, fmt.Errorf("decode WARP registration response: %w", err)
	}
	if len(parsed.Config.Peers) == 0 || strings.TrimSpace(parsed.Config.Peers[0].PublicKey) == "" {
		return profiles.RelayEndpoint{}, fmt.Errorf("WARP registration response missing peer public key")
	}

	peer := parsed.Config.Peers[0]
	endpoint := normalizeWARPEndpoint(peer.Endpoint.V4, peer.Endpoint.Host)
	addrV4 := strings.TrimSpace(parsed.Config.Interface.Addresses.V4)
	if addrV4 == "" {
		addrV4 = "172.16.0.2/32"
	} else if !strings.Contains(addrV4, "/") {
		addrV4 += "/32"
	}

	cfg := &profiles.WireGuardConfig{
		PrivateKey:          privB64,
		PublicKey:           strings.TrimSpace(peer.PublicKey),
		Address:             addrV4,
		DNS:                 "1.1.1.1",
		MTU:                 1280,
		PersistentKeepalive: 25,
	}
	if err := profiles.ValidateWireGuardConfig(cfg, endpoint); err != nil {
		return profiles.RelayEndpoint{}, fmt.Errorf("validate WARP config: %w", err)
	}

	return profiles.RelayEndpoint{
		ID:        "cloudflare-warp-free",
		Name:      "Cloudflare WARP (Free)",
		Continent: "Global",
		Location:  "Cloudflare Anycast",
		Endpoint:  endpoint,
		Tier:      "Free",
		Transport: "wireguard",
		WireGuard: cfg,
	}, nil
}

func normalizeWARPEndpoint(v4, host string) string {
	for _, candidate := range []string{strings.TrimSpace(v4), strings.TrimSpace(host)} {
		if candidate == "" {
			continue
		}
		h, p, err := net.SplitHostPort(candidate)
		if err == nil && h != "" {
			if p == "" || p == "0" {
				p = "2408"
			}
			return net.JoinHostPort(h, p)
		}
		if ip := net.ParseIP(candidate); ip != nil && ip.To4() != nil {
			return net.JoinHostPort(ip.String(), "2408")
		}
	}
	return defaultWARPFallbackEndpoint
}
