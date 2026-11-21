package profiles

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

// ParseWireGuardConfig parses a standard WireGuard / Cloudflare WARP INI config
// (wg-quick format) into a RelayEndpoint with Transport="wireguard".
//
// Only the client's first IPv4 address is selected because Lagvex's split-tunnel
// routing plane operates on IPv4 game CIDRs.
func ParseWireGuardConfig(raw string, name, location, continent string) (RelayEndpoint, error) {
	cfg := &WireGuardConfig{
		MTU:                 1280,
		PersistentKeepalive: 25,
	}
	var endpoint string

	section := ""
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if idx := strings.IndexAny(line, "#;"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
			if line == "" {
				continue
			}
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(k))
		val := strings.TrimSpace(v)

		switch section {
		case "interface":
			switch key {
			case "privatekey":
				cfg.PrivateKey = val
			case "address":
				if ip4 := firstIPv4Address(val); ip4 != "" {
					cfg.Address = ip4
				}
			case "dns":
				cfg.DNS = val
			case "mtu":
				if n, err := strconv.Atoi(val); err == nil && n >= 1200 && n <= 1500 {
					cfg.MTU = n
				}
			}
		case "peer":
			switch key {
			case "publickey":
				cfg.PublicKey = val
			case "presharedkey":
				cfg.PresharedKey = val
			case "endpoint":
				endpoint = val
			case "persistentkeepalive":
				if n, err := strconv.Atoi(val); err == nil && n >= 0 && n <= 600 {
					cfg.PersistentKeepalive = n
				}
			}
		}
	}

	if err := ValidateWireGuardConfig(cfg, endpoint); err != nil {
		return RelayEndpoint{}, err
	}

	host, port, _ := net.SplitHostPort(endpoint)
	name = strings.TrimSpace(name)
	if name == "" {
		name = "WireGuard (" + host + ")"
	}
	location = strings.TrimSpace(location)
	if location == "" {
		location = host
	}
	continent = strings.TrimSpace(continent)
	if continent == "" {
		continent = "Custom"
	}

	return RelayEndpoint{
		ID:        slugifyRelayID("wg-" + host + "-" + port),
		Name:      name,
		Continent: continent,
		Location:  location,
		Endpoint:  endpoint,
		Tier:      "WireGuard",
		Transport: "wireguard",
		WireGuard: cfg,
	}, nil
}

// ValidateWireGuardConfig verifies that all required WireGuard fields and keys
// are syntactically valid before saving or dialing.
func ValidateWireGuardConfig(cfg *WireGuardConfig, endpoint string) error {
	if cfg == nil {
		return fmt.Errorf("wireguard configuration is missing")
	}
	if _, err := DecodeWireGuardKeyHex(cfg.PrivateKey); err != nil {
		return fmt.Errorf("invalid WireGuard PrivateKey: %w", err)
	}
	if _, err := DecodeWireGuardKeyHex(cfg.PublicKey); err != nil {
		return fmt.Errorf("invalid WireGuard PublicKey: %w", err)
	}
	if strings.TrimSpace(cfg.PresharedKey) != "" {
		if _, err := DecodeWireGuardKeyHex(cfg.PresharedKey); err != nil {
			return fmt.Errorf("invalid WireGuard PresharedKey: %w", err)
		}
	}
	host, port, err := net.SplitHostPort(strings.TrimSpace(endpoint))
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("invalid WireGuard Endpoint %q (expected host:port)", endpoint)
	}
	if _, err := ParseWireGuardIPv4(cfg.Address); err != nil {
		return fmt.Errorf("invalid WireGuard Interface Address %q: %w", cfg.Address, err)
	}
	return nil
}

// DecodeWireGuardKeyHex decodes a base64 32-byte WireGuard key into lowercase
// hex as expected by wireguard-go's UAPI IpcSet format.
func DecodeWireGuardKeyHex(b64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return "", err
	}
	if len(raw) != 32 {
		return "", fmt.Errorf("expected 32-byte key, got %d bytes", len(raw))
	}
	return hex.EncodeToString(raw), nil
}

// ParseWireGuardIPv4 extracts the IPv4 address and prefix length from an
// Interface Address value such as "172.16.0.2/32" or "10.0.0.2".
func ParseWireGuardIPv4(addr string) (netip.Prefix, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return netip.Prefix{}, fmt.Errorf("address is empty")
	}
	if strings.Contains(addr, "/") {
		p, err := netip.ParsePrefix(addr)
		if err != nil {
			return netip.Prefix{}, err
		}
		if !p.Addr().Is4() {
			return netip.Prefix{}, fmt.Errorf("address must be IPv4")
		}
		return p, nil
	}
	ip, err := netip.ParseAddr(addr)
	if err != nil || !ip.Is4() {
		return netip.Prefix{}, fmt.Errorf("address must be valid IPv4")
	}
	return netip.PrefixFrom(ip, 32), nil
}

func firstIPv4Address(csv string) string {
	for _, part := range strings.Split(csv, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if p, err := ParseWireGuardIPv4(part); err == nil {
			return p.String()
		}
	}
	return ""
}
