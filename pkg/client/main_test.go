package client

import (
	"os"
	"testing"
)

// TestMain gives the whole client test binary its own config directory.
//
// The dashboard persists user config (custom games/relays, game-server history)
// in a machine-wide file. Without this, a dashboard running while tests run —
// or a previous test that recorded history — changes what later tests read
// back, and a suite that passes on a clean machine can fail on a used one.
// Same failure mode we hit live during round 20.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "lagvex-client-test-config")
	if err == nil {
		_ = os.Setenv("LAGVEX_CONFIG_DIR", dir)
	}
	code := m.Run()
	if dir != "" {
		_ = os.RemoveAll(dir)
	}
	os.Exit(code)
}
