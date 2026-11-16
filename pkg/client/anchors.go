package client

import (
	"strings"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// anchorProbeIDForRegion maps a game/region selection to the shared anchor
// allowlist id (protocol.AnchorProbeTargets), or 0 when the region has no
// anchor. It is used by both the dashboard handlers and the engine's standby
// pre-measurement.
func anchorProbeIDForRegion(pm *profiles.Manager, gameID, regionID string) byte {
	if pm == nil || gameID == "" || regionID == "" {
		return 0
	}
	game, ok := pm.FindGameByID(gameID)
	if !ok {
		return 0
	}
	regionName := ""
	for _, region := range game.Regions {
		if strings.EqualFold(region.ID, regionID) {
			regionName = region.Name
			break
		}
	}
	if regionName == "" {
		return 0
	}
	if anchor, ok := matchTelemetryAnchor(regionID, regionName); ok {
		return anchor.probeID
	}
	return 0
}
