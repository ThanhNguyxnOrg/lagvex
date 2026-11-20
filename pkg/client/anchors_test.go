package client

import (
	"path/filepath"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

// shippedProfilesPath points at the profile catalog that ships with the app.
func shippedProfilesPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "configs", "profiles.json")
}

func TestMatchTelemetryAnchorPrefersTheMostSpecificKey(t *testing.T) {
	cases := []struct {
		regionID string
		regionNm string
		wantID   byte
		wantLbl  string
	}{
		// Generic continental prefixes must not swallow a specific city.
		{"eu-london", "Europe West (London)", 13, "London"},
		{"eu-paris", "Europe West (Paris)", 12, "Paris"},
		{"eu-frankfurt", "Europe Central (Frankfurt)", 9, "Frankfurt"},
		{"euw-frankfurt", "Europe West (Frankfurt)", 9, "Frankfurt"},
		{"eu-stockholm", "Europe North (Stockholm)", 9, "Frankfurt"},
		{"eune-stockholm", "Europe North (Stockholm)", 9, "Frankfurt"},
		{"na-us-central", "North America Central (Dallas)", 16, "US Central"},
		{"na-us-east", "North America East (Virginia)", 7, "US East"},
		{"na-us-west", "North America West (Oregon)", 8, "US West"},
		{"asia-jp", "Japan (Tokyo)", 2, "Tokyo"},
		{"asia-tw", "Taiwan (EA Multiplay)", 4, "Hong Kong"},
		{"me-dxb", "Middle East (UAE North - Dubai)", 14, "Middle East"},
		{"af-jnb", "South Africa (South Africa North)", 11, "Johannesburg"},
		{"oc-syd", "Oceania (Sydney)", 6, "Sydney"},
		{"sa-br", "Brazil (São Paulo)", 10, "Sao Paulo"},
	}

	for _, tc := range cases {
		anchor, ok := matchTelemetryAnchor(tc.regionID, tc.regionNm)
		if !ok {
			t.Fatalf("region %q: no anchor matched", tc.regionID)
		}
		if anchor.probeID != tc.wantID {
			t.Errorf("region %q: anchor id = %d (%s), want %d (%s)",
				tc.regionID, anchor.probeID, anchor.label, tc.wantID, tc.wantLbl)
		}
		if anchor.label != tc.wantLbl {
			t.Errorf("region %q: anchor label = %q, want %q", tc.regionID, anchor.label, tc.wantLbl)
		}
		if _, ok := protocol.AnchorProbeTargets[anchor.probeID]; !ok {
			t.Errorf("region %q: anchor id %d has no relay-side target", tc.regionID, anchor.probeID)
		}
	}
}

func TestMatchTelemetryAnchorRejectsUnknownRegions(t *testing.T) {
	if _, ok := matchTelemetryAnchor("xx-nowhere", "Nowhere Land"); ok {
		t.Fatal("an unknown region must not resolve to an anchor")
	}
}

// TestShippedProfileRegionsResolveToAllowlistedAnchors walks every region in
// the shipped profile catalog and asserts that each resolved anchor points at
// a target the relay is actually allowed to probe. A stale mapping would
// otherwise send the relay to the wrong region (or nowhere at all).
func TestShippedProfileRegionsResolveToAllowlistedAnchors(t *testing.T) {
	pm, err := profiles.NewManager(shippedProfilesPath(t))
	if err != nil {
		t.Fatalf("load shipped profiles: %v", err)
	}

	resolved := 0
	for _, game := range pm.Catalog().Games {
		for _, region := range game.Regions {
			id := anchorProbeIDForRegion(pm, game.ID, region.ID)
			if id == 0 {
				continue
			}
			resolved++
			if _, ok := protocol.AnchorProbeTargets[id]; !ok {
				t.Errorf("game %q region %q resolves to anchor %d which is not allowlisted",
					game.ID, region.ID, id)
			}
		}
	}
	if resolved == 0 {
		t.Fatal("no shipped region resolved to an anchor — the mapping table is broken")
	}
	t.Logf("%d shipped game regions resolve to an allowlisted anchor", resolved)
}

// TestEveryAnchorHasARelayTarget keeps the client table and the relay
// allowlist in lockstep: an anchor without a relay-side target would make
// the dashboard offer a segment measurement the relay can never answer.
func TestEveryAnchorHasARelayTarget(t *testing.T) {
	for _, anchor := range telemetryAnchors {
		target, ok := protocol.AnchorProbeTargets[anchor.probeID]
		if !ok {
			t.Errorf("anchor %q (id %d) has no relay-side target", anchor.label, anchor.probeID)
			continue
		}
		if target == "" {
			t.Errorf("anchor %q (id %d) has an empty relay-side target", anchor.label, anchor.probeID)
		}
	}
	if len(telemetryAnchors) != len(protocol.AnchorProbeTargets) {
		t.Errorf("client anchors = %d, relay allowlist = %d — they must stay in sync",
			len(telemetryAnchors), len(protocol.AnchorProbeTargets))
	}
}
