package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// UserConfig represents the persistent settings and custom entries of the gamer.
type UserConfig struct {
	Nickname          string                        `json:"nickname"`
	LastSelectedGame  string                        `json:"lastSelectedGame"`
	LastSelectedRelay string                        `json:"lastSelectedRelay"`
	CustomGames       []profiles.GameDefinition     `json:"customGames"`
	CustomRelays      []profiles.RelayEndpoint      `json:"customRelays"`
	Tweaks            map[string]bool               `json:"tweaks"`
	RecentRelays      []string                      `json:"recentRelays,omitempty"`
	GameServers       map[string][]GameServerRecord `json:"gameServers,omitempty"`
}

// GameServerRecord remembers one measured game-server endpoint (GearUP keeps a
// per-game server history so a returning player sees the endpoints their
// sessions actually used, with the last measured RTT next to them).
type GameServerRecord struct {
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	AvgMs    float64   `json:"avgMs"`
	LossPct  float64   `json:"lossPct"`
	Samples  int       `json:"samples"`
	LastSeen time.Time `json:"lastSeen"`
}

// gameServerHistoryCap bounds the retained endpoints per game.
const gameServerHistoryCap = 8

// ConfigStore manages reading and writing UserConfig to disk atomically.
type ConfigStore struct {
	mu       sync.RWMutex
	filePath string
	config   UserConfig
}

// NewConfigStore creates a new config store. It automatically resolves the AppData directory.
func NewConfigStore(customPath string) *ConfigStore {
	path := customPath
	if path == "" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, ".config")
		}
		dir := filepath.Join(appData, "Lagvex")
		_ = os.MkdirAll(dir, 0755)
		path = filepath.Join(dir, "user_config.json")
	}

	cs := &ConfigStore{
		filePath: path,
		config: UserConfig{
			Nickname:     "Gamer",
			CustomGames:  make([]profiles.GameDefinition, 0),
			CustomRelays: make([]profiles.RelayEndpoint, 0),
			Tweaks: map[string]bool{
				"tcpNoDelay":    true,
				"disableNagle":  true,
				"mmcssPriority": true,
				"mtuClamping":   true,
				"smartL4Filter": true,
			},
		},
	}
	cs.load()
	return cs
}

func (cs *ConfigStore) load() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &cs.config)
}

// Save writes current config to disk.
func (cs *ConfigStore) Save() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	data, err := json.MarshalIndent(cs.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.filePath, data, 0644)
}

// Get returns a copy of current UserConfig.
func (cs *ConfigStore) Get() UserConfig {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.config
}

// SetNickname updates the gamer nickname and persists to disk.
func (cs *ConfigStore) SetNickname(nick string) {
	cs.mu.Lock()
	cs.config.Nickname = nick
	cs.mu.Unlock()
	_ = cs.Save()
}

// AddCustomGame adds a custom game and persists to disk.
func (cs *ConfigStore) AddCustomGame(game profiles.GameDefinition) {
	cs.mu.Lock()
	// Replace if exists
	found := false
	for i, g := range cs.config.CustomGames {
		if g.ID == game.ID {
			cs.config.CustomGames[i] = game
			found = true
			break
		}
	}
	if !found {
		cs.config.CustomGames = append(cs.config.CustomGames, game)
	}
	cs.mu.Unlock()
	_ = cs.Save()
}

// AddCustomRelay adds a custom relay node and persists to disk.
func (cs *ConfigStore) AddCustomRelay(relay profiles.RelayEndpoint) {
	cs.mu.Lock()
	found := false
	for i, r := range cs.config.CustomRelays {
		if r.ID == relay.ID {
			cs.config.CustomRelays[i] = relay
			found = true
			break
		}
	}
	if !found {
		cs.config.CustomRelays = append(cs.config.CustomRelays, relay)
	}
	cs.mu.Unlock()
	_ = cs.Save()
}

// AddRecentRelay remembers the last few relays the user connected through,
// most recent first (max 3), for quick re-selection like the node-history
// lists in GearUP and Mudfish.
func (cs *ConfigStore) AddRecentRelay(endpoint string) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return
	}

	cs.mu.Lock()
	recent := make([]string, 0, 4)
	recent = append(recent, endpoint)
	for _, existing := range cs.config.RecentRelays {
		if len(recent) >= 3 {
			break
		}
		if !strings.EqualFold(existing, endpoint) {
			recent = append(recent, existing)
		}
	}
	cs.config.RecentRelays = recent
	cs.mu.Unlock()
	_ = cs.Save()
}

// SetSelections persists the last selected game and relay ids.
func (cs *ConfigStore) SetSelections(gameID, relayID string) {
	cs.mu.Lock()
	if gameID != "" {
		cs.config.LastSelectedGame = gameID
	}
	if relayID != "" {
		cs.config.LastSelectedRelay = relayID
	}
	cs.mu.Unlock()
	_ = cs.Save()
}

// SetTweaks updates system tweak flags and persists to disk.
func (cs *ConfigStore) SetTweaks(tweaks map[string]bool) {
	cs.mu.Lock()
	if cs.config.Tweaks == nil {
		cs.config.Tweaks = make(map[string]bool)
	}
	for k, v := range tweaks {
		cs.config.Tweaks[k] = v
	}
	cs.mu.Unlock()
	_ = cs.Save()
}

// RecordGameServer stores one measured game-server endpoint for a game, most
// recent first, capped per game. Only real measurements belong here: callers
// must pass a positive AvgMs, so a failed probe can never fabricate history.
func (cs *ConfigStore) RecordGameServer(gameID string, rec GameServerRecord) {
	gameID = strings.TrimSpace(gameID)
	rec.Host = strings.TrimSpace(rec.Host)
	if gameID == "" || rec.Host == "" || rec.AvgMs <= 0 {
		return
	}

	cs.mu.Lock()
	if cs.config.GameServers == nil {
		cs.config.GameServers = make(map[string][]GameServerRecord)
	}
	if rec.LastSeen.IsZero() {
		rec.LastSeen = time.Now().UTC()
	}
	if rec.Port <= 0 || rec.Port > 65535 {
		rec.Port = 443
	}
	if rec.Samples <= 0 {
		rec.Samples = 1
	}
	list := make([]GameServerRecord, 0, gameServerHistoryCap)
	list = append(list, rec)
	for _, existing := range cs.config.GameServers[gameID] {
		if strings.EqualFold(existing.Host, rec.Host) && existing.Port == rec.Port {
			continue
		}
		if len(list) >= gameServerHistoryCap {
			break
		}
		list = append(list, existing)
	}
	cs.config.GameServers[gameID] = list
	cs.mu.Unlock()
	_ = cs.Save()
}

// GameServersFor returns a copy of the persisted game-server history for a game.
func (cs *ConfigStore) GameServersFor(gameID string) []GameServerRecord {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	list := cs.config.GameServers[gameID]
	if len(list) == 0 {
		return nil
	}
	out := make([]GameServerRecord, len(list))
	copy(out, list)
	return out
}
