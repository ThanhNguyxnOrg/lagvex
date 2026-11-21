package profiles

import (
	"time"
)

// GameRegion defines a specific game server region with its associated CIDR IP ranges.
type GameRegion struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Source string   `json:"source,omitempty"`
	CIDRs  []string `json:"cidrs"`
}

// GameDefinition describes a game, its executables, and supported server regions.
type GameDefinition struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Category       string       `json:"category"`
	Icon           string       `json:"icon,omitempty"`
	ProcessNames   []string     `json:"processNames"`
	LobbyAddresses []string     `json:"lobbyAddresses,omitempty"`
	Regions        []GameRegion `json:"regions"`
}

// WireGuardConfig stores the configuration for a WireGuard / Cloudflare WARP
// transport endpoint. Secrets (PrivateKey, PresharedKey) are stripped before
// exposing relays to the dashboard or diagnostic exports.
type WireGuardConfig struct {
	PrivateKey          string `json:"privateKey,omitempty"`
	PublicKey           string `json:"publicKey"`
	PresharedKey        string `json:"presharedKey,omitempty"`
	Address             string `json:"address"`
	DNS                 string `json:"dns,omitempty"`
	MTU                 int    `json:"mtu,omitempty"`
	PersistentKeepalive int    `json:"persistentKeepalive,omitempty"`
}

// RelayEndpoint describes a Lagvex relay server or a WireGuard/WARP endpoint.
type RelayEndpoint struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Continent string           `json:"continent,omitempty"`
	Location  string           `json:"location"`
	Endpoint  string           `json:"endpoint"` // host:port
	PSK       string           `json:"psk,omitempty"`
	Tier      string           `json:"tier,omitempty"`
	Transport string           `json:"transport,omitempty"` // "lagvex" (default) or "wireguard"
	WireGuard *WireGuardConfig `json:"wireguard,omitempty"`
}

// IsWireGuard reports whether this relay uses the WireGuard transport.
func (r RelayEndpoint) IsWireGuard() bool {
	return r.Transport == "wireguard" || r.WireGuard != nil
}

// Redacted returns a copy of the relay with secret credentials removed.
func (r RelayEndpoint) Redacted() RelayEndpoint {
	r.PSK = ""
	if r.WireGuard != nil {
		wgCopy := *r.WireGuard
		wgCopy.PrivateKey = ""
		wgCopy.PresharedKey = ""
		r.WireGuard = &wgCopy
	}
	return r
}

// ProfileCatalog contains all games and known relays.
type ProfileCatalog struct {
	SchemaVersion int              `json:"schemaVersion"`
	UpdatedUtc    time.Time        `json:"updatedUtc"`
	Games         []GameDefinition `json:"games"`
	Relays        []RelayEndpoint  `json:"relays"`
}
