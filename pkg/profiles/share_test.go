package profiles

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRelayShareURIRoundTrip(t *testing.T) {
	orig := RelayEndpoint{
		ID:        "sg-01",
		Name:      "Singapore Vultr",
		Continent: "Asia-Pacific",
		Location:  "Singapore",
		Endpoint:  "203.0.113.10:51820",
		PSK:       "super-secret-key-123",
		Tier:      "Squad",
	}

	uri, err := EncodeRelayShareURI(orig)
	if err != nil {
		t.Fatalf("EncodeRelayShareURI failed: %v", err)
	}
	if !strings.HasPrefix(uri, "lagvex://203.0.113.10:51820?") {
		t.Fatalf("unexpected URI prefix: %s", uri)
	}

	parsed, err := ParseRelayShareURI(uri)
	if err != nil {
		t.Fatalf("ParseRelayShareURI failed: %v", err)
	}
	if parsed.Endpoint != orig.Endpoint || parsed.PSK != orig.PSK || parsed.Name != orig.Name || parsed.Location != orig.Location || parsed.Continent != orig.Continent || parsed.ID != orig.ID {
		t.Fatalf("round-trip mismatch:\ngot:  %+v\nwant: %+v", parsed, orig)
	}
}

func TestParseRelayShareURIErrors(t *testing.T) {
	cases := []string{
		"",
		"https://example.com:51820?psk=abc",
		"lagvex://203.0.113.10:51820",
		"lagvex://invalid-no-port?psk=abc",
	}
	for _, tc := range cases {
		if _, err := ParseRelayShareURI(tc); err == nil {
			t.Errorf("expected error for %q, got nil", tc)
		}
	}
}

func TestParseWireGuardConfig(t *testing.T) {
	priv := base64.StdEncoding.EncodeToString(make([]byte, 32))
	pub := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

	conf := `
# Sample Cloudflare WARP WireGuard config
[Interface]
PrivateKey = ` + priv + `
Address = 172.16.0.2/32, 2606:4700:110::2/128
DNS = 1.1.1.1
MTU = 1280

[Peer]
PublicKey = ` + pub + `
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = 162.159.192.1:2408
PersistentKeepalive = 25
`
	rel, err := ParseWireGuardConfig(conf, "Cloudflare WARP Free", "Anycast", "Global")
	if err != nil {
		t.Fatalf("ParseWireGuardConfig failed: %v", err)
	}
	if !rel.IsWireGuard() {
		t.Fatalf("expected IsWireGuard()=true")
	}
	if rel.Endpoint != "162.159.192.1:2408" {
		t.Fatalf("endpoint = %q, want 162.159.192.1:2408", rel.Endpoint)
	}
	if rel.WireGuard.Address != "172.16.0.2/32" {
		t.Fatalf("address = %q, want 172.16.0.2/32", rel.WireGuard.Address)
	}
	if rel.WireGuard.MTU != 1280 {
		t.Fatalf("mtu = %d, want 1280", rel.WireGuard.MTU)
	}

	redacted := rel.Redacted()
	if redacted.WireGuard.PrivateKey != "" {
		t.Fatalf("expected Redacted() to clear PrivateKey")
	}
	if redacted.WireGuard.PublicKey != pub {
		t.Fatalf("expected Redacted() to keep PublicKey")
	}
}
