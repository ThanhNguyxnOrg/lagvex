package profiles

import (
	"net/netip"
	"path/filepath"
	"strings"
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

func TestAllCIDRsValidAndNoBogonCollision(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "profiles.json")
	mgr, err := NewManager(path)
	if err != nil {
		t.Fatalf("failed to load profiles: %v", err)
	}

	catalog := mgr.Catalog()
	totalCIDRs := 0

	for _, game := range catalog.Games {
		for _, reg := range game.Regions {
			for _, cidr := range reg.CIDRs {
				totalCIDRs++
				prefix, err := netip.ParsePrefix(cidr)
				if err != nil {
					t.Errorf("game %s, region %s has invalid CIDR %q: %v", game.ID, reg.ID, cidr, err)
					continue
				}

				addr := prefix.Addr()
				if !addr.Is4() {
					t.Errorf("game %s, region %s has non-IPv4 CIDR %q", game.ID, reg.ID, cidr)
					continue
				}

				// Check bogon/private collision
				if addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsMulticast() {
					t.Errorf("game %s, region %s has loopback/link-local/multicast CIDR %q", game.ID, reg.ID, cidr)
				}
			}
		}
	}

	t.Logf("Validated %d total game CIDRs successfully", totalCIDRs)
}

func TestEmbeddedDefaultsFallback(t *testing.T) {
	// Call NewManager with empty string and non-existent path to verify embedded fallback
	mgr, err := NewManager("")
	if err != nil {
		t.Fatalf("failed to init manager with embedded fallback: %v", err)
	}

	catalog := mgr.Catalog()
	if len(catalog.Games) < 10 {
		t.Errorf("expected at least 10 embedded games, got %d", len(catalog.Games))
	}

	// Lagvex does not operate public relays. Shipping catalog entries that
	// point at dead or third-party infrastructure makes the UI claim nodes are
	// available when they are not, so the embedded catalog must stay empty.
	if len(catalog.Relays) != 0 {
		t.Errorf("embedded catalog must not bundle relay nodes, got %d: %+v", len(catalog.Relays), catalog.Relays)
	}
	for _, r := range catalog.Relays {
		if strings.Contains(strings.ToLower(r.Endpoint), "lagvex.org") {
			t.Errorf("embedded catalog contains a retired lagvex.org relay: %+v", r)
		}
	}

	// Verify Valorant exists in embedded
	val, ok := mgr.FindGameByID("valorant")
	if !ok || val.Name != "Valorant" {
		t.Errorf("expected Valorant in embedded catalog, got %+v", val)
	}
}
