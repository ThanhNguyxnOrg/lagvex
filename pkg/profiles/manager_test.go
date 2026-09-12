package profiles

import (
	"path/filepath"
	"testing"
)

func TestProfilesLoading(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "profiles.json")
	mgr, err := NewManager(path)
	if err != nil {
		t.Fatalf("failed to load profiles: %v", err)
	}

	catalog := mgr.Catalog()
	if len(catalog.Games) == 0 {
		t.Fatalf("expected games in catalog, got 0")
	}

	// Verify key games exist
	gamesToCheck := []string{"valorant", "cs2", "pubg", "apex", "thefinals", "cod_warzone", "lol", "dota2"}
	for _, gid := range gamesToCheck {
		game, ok := mgr.FindGameByID(gid)
		if !ok {
			t.Errorf("game %s not found in catalog", gid)
			continue
		}
		if len(game.Regions) == 0 {
			t.Errorf("game %s has no regions configured", gid)
		}
		if len(game.ProcessNames) == 0 {
			t.Errorf("game %s has no process names configured", gid)
		}
	}

	// Test process lookup
	game, ok := mgr.FindGameByProcess("VALORANT-Win64-Shipping.exe")
	if !ok || game.ID != "valorant" {
		t.Fatalf("expected valorant by process, got %+v", game)
	}

	cs2Game, ok := mgr.FindGameByProcess("cs2.exe")
	if !ok || cs2Game.ID != "cs2" {
		t.Fatalf("expected cs2 by process, got %+v", cs2Game)
	}

	// Test CIDR retrieval
	cidrs, err := mgr.GetRegionCIDRs("valorant", "asia-sg")
	if err != nil || len(cidrs) == 0 {
		t.Fatalf("expected valorant asia-sg cidrs, got err: %v, len: %d", err, len(cidrs))
	}
}
