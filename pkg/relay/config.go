package relay

import (
	"errors"
	"net/netip"
	"time"
)

// Config holds all parameters required by the Relay server.
type Config struct {
	ListenAddr  string        // UDP address to bind, e.g. ":51820"
	HTTPAddr    string        // HTTP address for squad signaling & status, e.g. ":51821"
	TunName     string        // Linux TUN device name, e.g. "lagvex0"
	Subnet      netip.Prefix  // Client subnet, e.g. "10.88.0.0/24"
	PSK         []byte        // Pre-shared key for authentication
	MTU         int           // MTU for the tunnel interface, e.g. 1400
	IdleTimeout time.Duration // Maximum inactivity before session cleanup (e.g. 90s)
	MaxClients  int           // Maximum concurrent sessions (0 = pool size)
}

// Validate checks configuration sanity.
func (c *Config) Validate() error {
	if c.ListenAddr == "" {
		return errors.New("listen address is required")
	}
	if c.TunName == "" {
		c.TunName = "lagvex0"
	}
	if !c.Subnet.IsValid() || !c.Subnet.Addr().Is4() {
		return errors.New("valid IPv4 subnet prefix is required")
	}
	if len(c.PSK) == 0 {
		return errors.New("pre-shared key (PSK) cannot be empty")
	}
	if c.MTU <= 0 {
		c.MTU = 1400
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 90 * time.Second
	}
	return nil
}
