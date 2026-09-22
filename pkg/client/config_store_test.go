package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

func TestConfigStorePersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lagvex_cfg_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfgPath := filepath.Join(tempDir, "user_config.json")
	cs := NewConfigStore(cfgPath)

	// 1. Set nickname
	cs.SetNickname("CyberSniper#9999")
	if cs.Get().Nickname != "CyberSniper#9999" {
		t.Fatalf("expected nickname CyberSniper#9999, got %s", cs.Get().Nickname)
	}

	// 2. Add custom game
	customGame := profiles.GameDefinition{
		ID:           "testgame",
		Name:         "Test Shooter",
		Category:     "FPS",
		ProcessNames: []string{"testshooter.exe"},
		Regions: []profiles.GameRegion{
			{ID: "asia", Name: "Asia-Pacific", CIDRs: []string{"1.1.1.0/24"}},
		},
	}
	cs.AddCustomGame(customGame)
	if len(cs.Get().CustomGames) != 1 || cs.Get().CustomGames[0].ID != "testgame" {
		t.Fatalf("expected 1 custom game with id testgame, got %+v", cs.Get().CustomGames)
	}

	// 3. Add custom relay
	customRelay := profiles.RelayEndpoint{
		ID:       "tokyo-test",
		Name:     "Tokyo Test Node",
		Endpoint: "192.0.2.1:51820",
		Location: "Tokyo, Japan",
		Tier:     "custom",
	}
	cs.AddCustomRelay(customRelay)
	if len(cs.Get().CustomRelays) != 1 || cs.Get().CustomRelays[0].ID != "tokyo-test" {
		t.Fatalf("expected 1 custom relay, got %+v", cs.Get().CustomRelays)
	}

	// 4. Update Tweaks
	cs.SetTweaks(map[string]bool{
		"tcpNoDelay":   false,
		"disableNagle": true,
	})
	if cs.Get().Tweaks["tcpNoDelay"] != false || cs.Get().Tweaks["disableNagle"] != true {
		t.Fatalf("tweaks were not properly updated: %+v", cs.Get().Tweaks)
	}

	// 5. Create a second instance pointing to the same path to verify reload
	cs2 := NewConfigStore(cfgPath)
	cfg2 := cs2.Get()
	if cfg2.Nickname != "CyberSniper#9999" {
		t.Fatalf("expected reloaded nickname CyberSniper#9999, got %s", cfg2.Nickname)
	}
	if len(cfg2.CustomGames) != 1 || cfg2.CustomGames[0].Name != "Test Shooter" {
		t.Fatalf("expected reloaded custom game Test Shooter, got %+v", cfg2.CustomGames)
	}
	if len(cfg2.CustomRelays) != 1 || cfg2.CustomRelays[0].Endpoint != "192.0.2.1:51820" {
		t.Fatalf("expected reloaded custom relay, got %+v", cfg2.CustomRelays)
	}
}
