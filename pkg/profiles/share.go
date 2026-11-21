package profiles

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// EncodeRelayShareURI serializes a native Lagvex relay into a portable
// lagvex:// share URI so a squad host can share one VPS configuration with a
// single copy/paste.
//
// Format:
//
//	lagvex://<host>:<port>?psk=<psk>&name=<name>&loc=<location>&continent=<continent>&id=<id>
func EncodeRelayShareURI(relay RelayEndpoint) (string, error) {
	if relay.IsWireGuard() {
		return "", fmt.Errorf("wireguard endpoints cannot be encoded as a lagvex:// URI")
	}
	ep := strings.TrimSpace(relay.Endpoint)
	host, port, err := net.SplitHostPort(ep)
	if err != nil || host == "" || port == "" {
		return "", fmt.Errorf("invalid relay endpoint %q (expected host:port)", relay.Endpoint)
	}
	if strings.TrimSpace(relay.PSK) == "" {
		return "", fmt.Errorf("relay PSK is required to generate a share URI")
	}

	q := url.Values{}
	q.Set("psk", relay.PSK)
	if strings.TrimSpace(relay.Name) != "" {
		q.Set("name", strings.TrimSpace(relay.Name))
	}
	if strings.TrimSpace(relay.Location) != "" {
		q.Set("loc", strings.TrimSpace(relay.Location))
	}
	if strings.TrimSpace(relay.Continent) != "" {
		q.Set("continent", strings.TrimSpace(relay.Continent))
	}
	if strings.TrimSpace(relay.ID) != "" {
		q.Set("id", strings.TrimSpace(relay.ID))
	}
	if strings.TrimSpace(relay.Tier) != "" {
		q.Set("tier", strings.TrimSpace(relay.Tier))
	}

	u := &url.URL{
		Scheme:   "lagvex",
		Host:     net.JoinHostPort(host, port),
		RawQuery: q.Encode(),
	}
	return u.String(), nil
}

// ParseRelayShareURI parses a lagvex:// share URI into a validated RelayEndpoint.
func ParseRelayShareURI(raw string) (RelayEndpoint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RelayEndpoint{}, fmt.Errorf("share URI is empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return RelayEndpoint{}, fmt.Errorf("invalid share URI: %w", err)
	}
	if !strings.EqualFold(u.Scheme, "lagvex") {
		return RelayEndpoint{}, fmt.Errorf("unsupported URI scheme %q (expected lagvex://)", u.Scheme)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil || host == "" || port == "" {
		return RelayEndpoint{}, fmt.Errorf("invalid relay endpoint %q in share URI", u.Host)
	}

	q := u.Query()
	psk := strings.TrimSpace(q.Get("psk"))
	if psk == "" {
		return RelayEndpoint{}, fmt.Errorf("share URI is missing required psk parameter")
	}

	endpoint := net.JoinHostPort(host, port)
	name := strings.TrimSpace(q.Get("name"))
	if name == "" {
		name = "Shared Relay (" + host + ")"
	}
	location := strings.TrimSpace(q.Get("loc"))
	if location == "" {
		location = strings.TrimSpace(q.Get("location"))
	}
	if location == "" {
		location = host
	}
	continent := strings.TrimSpace(q.Get("continent"))
	if continent == "" {
		continent = "Custom"
	}
	id := strings.TrimSpace(q.Get("id"))
	if id == "" {
		id = slugifyRelayID("shared-" + host + "-" + port)
	}
	tier := strings.TrimSpace(q.Get("tier"))
	if tier == "" {
		tier = "Shared"
	}

	return RelayEndpoint{
		ID:        id,
		Name:      name,
		Continent: continent,
		Location:  location,
		Endpoint:  endpoint,
		PSK:       psk,
		Tier:      tier,
		Transport: "lagvex",
	}, nil
}

func slugifyRelayID(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.' || r == ':' || r == ' ':
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "relay-custom"
	}
	return out
}
