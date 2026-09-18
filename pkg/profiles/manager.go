package profiles

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed embedded/profiles.json
var defaultProfilesJSON []byte

//go:embed embedded/relays.json
var defaultRelaysJSON []byte

// Manager manages game profiles and relay catalogs.
type Manager struct {
	mu      sync.RWMutex
	catalog ProfileCatalog
	path    string
}

// NewManager loads the profile catalog from a file or falls back to the embedded default catalog.
func NewManager(filePath string) (*Manager, error) {
	m := &Manager{
		path: filePath,
	}

	loadedFromDisk := false
	if filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			if err := json.Unmarshal(data, &m.catalog); err == nil {
				loadedFromDisk = true
			}
		}
	}

	// If not loaded from disk or empty file path, populate with embedded default catalog
	if !loadedFromDisk {
		if err := json.Unmarshal(defaultProfilesJSON, &m.catalog); err != nil {
			m.catalog = ProfileCatalog{
				SchemaVersion: 1,
				Games:         []GameDefinition{},
				Relays:        []RelayEndpoint{},
			}
		}
	}

	// If catalog has no relays, attempt to load from sibling relays.json on disk
	if len(m.catalog.Relays) == 0 && filePath != "" {
		relaysPath := filepath.Join(filepath.Dir(filePath), "relays.json")
		if rData, err := os.ReadFile(relaysPath); err == nil {
			var relays []RelayEndpoint
			if err := json.Unmarshal(rData, &relays); err == nil {
				m.catalog.Relays = relays
			}
		}
	}

	// If still no relays, populate with embedded default community relays
	if len(m.catalog.Relays) == 0 && len(defaultRelaysJSON) > 0 {
		var relays []RelayEndpoint
		if err := json.Unmarshal(defaultRelaysJSON, &relays); err == nil {
			m.catalog.Relays = relays
		}
	}

	return m, nil
}

// Catalog returns a copy of the loaded catalog.
func (m *Manager) Catalog() ProfileCatalog {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.catalog
}

// FindGameByID looks up a game by ID.
func (m *Manager) FindGameByID(id string) (*GameDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, g := range m.catalog.Games {
		if strings.EqualFold(g.ID, id) {
			copied := g
			return &copied, true
		}
	}
	return nil, false
}

// FindGameByProcess matches a running executable name against registered game processes.
func (m *Manager) FindGameByProcess(procName string) (*GameDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	procLower := strings.ToLower(strings.TrimSpace(procName))
	for _, g := range m.catalog.Games {
		for _, p := range g.ProcessNames {
			if strings.ToLower(p) == procLower {
				copied := g
				return &copied, true
			}
		}
	}
	return nil, false
}

// GetRegionCIDRs returns verified CIDRs for a specific game and region.
func (m *Manager) GetRegionCIDRs(gameID, regionID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, ok := m.FindGameByID(gameID)
	if !ok {
		return nil, fmt.Errorf("game %q not found", gameID)
	}

	for _, reg := range game.Regions {
		if strings.EqualFold(reg.ID, regionID) {
			valid := make([]string, 0, len(reg.CIDRs))
			for _, cidr := range reg.CIDRs {
				if _, err := netip.ParsePrefix(cidr); err == nil {
					valid = append(valid, cidr)
				}
			}
			return valid, nil
		}
	}

	return nil, fmt.Errorf("region %q not found for game %q", regionID, gameID)
}

// AddGame adds or updates a game definition.
func (m *Manager) AddGame(game GameDefinition) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, existing := range m.catalog.Games {
		if strings.EqualFold(existing.ID, game.ID) {
			m.catalog.Games[i] = game
			return m.saveLocked()
		}
	}

	m.catalog.Games = append(m.catalog.Games, game)
	return m.saveLocked()
}

// AddRelay adds a relay server.
func (m *Manager) AddRelay(relay RelayEndpoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, existing := range m.catalog.Relays {
		if strings.EqualFold(existing.ID, relay.ID) {
			m.catalog.Relays[i] = relay
			return m.saveLocked()
		}
	}

	m.catalog.Relays = append(m.catalog.Relays, relay)
	return m.saveLocked()
}

func (m *Manager) saveLocked() error {
	if m.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(m.catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0644)
}
