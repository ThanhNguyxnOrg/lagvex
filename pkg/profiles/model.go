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

// RelayEndpoint describes a Lagvex relay server.
type RelayEndpoint struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Continent string `json:"continent,omitempty"`
	Location  string `json:"location"`
	Endpoint  string `json:"endpoint"` // host:port
	PSK       string `json:"psk,omitempty"`
	Tier      string `json:"tier,omitempty"`
}

// ProfileCatalog contains all games and known relays.
type ProfileCatalog struct {
	SchemaVersion int              `json:"schemaVersion"`
	UpdatedUtc    time.Time        `json:"updatedUtc"`
	Games         []GameDefinition `json:"games"`
	Relays        []RelayEndpoint  `json:"relays"`
}
