package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// UserConfig represents the persistent settings and custom entries of the gamer.
type UserConfig struct {
	Nickname          string                    `json:"nickname"`
	LastSelectedGame  string                    `json:"lastSelectedGame"`
	LastSelectedRelay string                    `json:"lastSelectedRelay"`
	CustomGames       []profiles.GameDefinition `json:"customGames"`
	CustomRelays      []profiles.RelayEndpoint  `json:"customRelays"`
	Tweaks            map[string]bool           `json:"tweaks"`
}

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
	cs.mu.RLock()
	defer cs.mu.RUnlock()

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
