package client

import (
	"testing"
	"time"
)

func TestFailoverDegradationDetection(t *testing.T) {
	fc := NewFailoverController(DefaultFailoverConfig())

	// Healthy scenario
	degraded, reason := fc.IsDegraded(0, 30.0, 30.0, 0.0)
	if degraded {
		t.Fatalf("expected healthy state, got degraded: %s", reason)
	}

	// Missed pongs degradation
	degraded, _ = fc.IsDegraded(3, 30.0, 30.0, 0.0)
	if !degraded {
		t.Fatalf("expected degraded on 3 missed pongs, got healthy")
	}

	// Packet loss degradation (threshold is 3.0%)
	degraded, _ = fc.IsDegraded(0, 30.0, 30.0, 3.5)
	if !degraded {
		t.Fatalf("expected degraded on 3.5%% loss, got healthy")
	}

	// Latency spike degradation (baseline 30ms -> current 50ms, +66% > 40%)
	degraded, _ = fc.IsDegraded(0, 50.0, 30.0, 0.0)
	if !degraded {
		t.Fatalf("expected degraded on latency spike (30ms -> 50ms), got healthy")
	}

	// Minor latency increase should NOT trigger degradation (30ms -> 35ms is < 40%)
	degraded, reason = fc.IsDegraded(0, 35.0, 30.0, 0.0)
	if degraded {
		t.Fatalf("did not expect degradation on minor +5ms increase, got: %s", reason)
	}
}

func TestFailoverHysteresisThreshold(t *testing.T) {
	cfg := DefaultFailoverConfig()
	cfg.MinScoreImprovement = 0.15 // 15% improvement required
	cfg.CooldownDuration = 5 * time.Second
	fc := NewFailoverController(cfg)

	now := time.Now()

	// Current score: 100. Target threshold = 100 * 0.85 = 85.
	// Candidate score: 90 (only 10% better) -> REJECT
	should, reason := fc.ShouldSwitch(100.0, 90.0, now)
	if should {
		t.Fatalf("expected rejection for 10%% improvement, got switch: %s", reason)
	}

	// Candidate score: 85 (exactly 15% better) -> ACCEPT
	should, reason = fc.ShouldSwitch(100.0, 85.0, now)
	if !should {
		t.Fatalf("expected acceptance for 15%% improvement, got rejected: %s", reason)
	}

	// Candidate score: 70 (30% better) -> ACCEPT
	should, reason = fc.ShouldSwitch(100.0, 70.0, now)
	if !should {
		t.Fatalf("expected acceptance for 30%% improvement, got rejected: %s", reason)
	}

	// If current node is dead/unreachable (score > 999999) -> ACCEPT immediately
	should, reason = fc.ShouldSwitch(1000000.0, 200.0, now)
	if !should {
		t.Fatalf("expected immediate switch when current node is dead, got rejected: %s", reason)
	}
}

func TestFailoverCooldownSuppression(t *testing.T) {
	cfg := DefaultFailoverConfig()
	cfg.CooldownDuration = 5 * time.Second
	fc := NewFailoverController(cfg)

	t0 := time.Now()

	// First switch: accepted
	should, _ := fc.ShouldSwitch(100.0, 70.0, t0)
	if !should {
		t.Fatalf("expected first switch to be accepted")
	}

	// Record switch at t0
	fc.RecordSwitch(FailoverEvent{
		Timestamp:  t0,
		OldRelayID: "relay-1",
		NewRelayID: "relay-2",
		OldScore:   100.0,
		NewScore:   70.0,
		HandoverOk: true,
	})

	// Try switching again 2 seconds later (within 5s cooldown) -> REJECT
	t1 := t0.Add(2 * time.Second)
	should, reason := fc.ShouldSwitch(70.0, 40.0, t1)
	if should {
		t.Fatalf("expected cooldown suppression at +2s, got switch: %s", reason)
	}

	// Try switching 6 seconds later (after 5s cooldown) -> ACCEPT
	t2 := t0.Add(6 * time.Second)
	should, reason = fc.ShouldSwitch(70.0, 40.0, t2)
	if !should {
		t.Fatalf("expected switch accepted after cooldown expiry, got: %s", reason)
	}

	// Verify history size
	history := fc.History()
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if history[0].OldRelayID != "relay-1" || history[0].NewRelayID != "relay-2" {
		t.Fatalf("unexpected history content: %+v", history[0])
	}
}

func TestEngineFailoverControls(t *testing.T) {
	eng, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}

	// Default should be enabled
	if !eng.IsAutoFailoverEnabled() {
		t.Fatalf("expected auto-failover to be enabled by default")
	}

	eng.SetAutoFailover(false)
	if eng.IsAutoFailoverEnabled() {
		t.Fatalf("expected auto-failover to be disabled after SetAutoFailover(false)")
	}

	eng.SetAutoFailover(true)
	if !eng.IsAutoFailoverEnabled() {
		t.Fatalf("expected auto-failover to be enabled after SetAutoFailover(true)")
	}

	if eng.FailoverController() == nil {
		t.Fatalf("expected non-nil FailoverController")
	}

	// Initial history should be empty
	hist := eng.FailoverHistory()
	if len(hist) != 0 {
		t.Fatalf("expected 0 initial history, got %d", len(hist))
	}
}

func TestFailoverResetHistory(t *testing.T) {
	fc := NewFailoverController(DefaultFailoverConfig())
	fc.RecordSwitch(FailoverEvent{Timestamp: time.Now(), HandoverOk: true})

	if len(fc.History()) != 1 {
		t.Fatalf("expected one recorded event, got %d", len(fc.History()))
	}
	if fc.LastSwitchTime().IsZero() {
		t.Fatalf("expected the cooldown timer to be set")
	}

	fc.ResetHistory()

	if len(fc.History()) != 0 {
		t.Fatalf("history must be empty after reset, got %d", len(fc.History()))
	}
	if !fc.LastSwitchTime().IsZero() {
		t.Fatalf("cooldown timer must be cleared after reset")
	}
}

func TestSelectStandbyResult(t *testing.T) {
	results := []ProbeResult{
		{RelayID: "active", Endpoint: "a:1", Reachable: true, Score: 5},
		{RelayID: "unstable", Endpoint: "b:1", Reachable: true, Score: 6, Unstable: true},
		{RelayID: "down", Endpoint: "c:1", Reachable: false},
		{RelayID: "good", Endpoint: "d:1", Reachable: true, Score: 20, RTTMedianMs: 20},
	}

	best, ok := selectStandbyResult(results, "a:1")
	if !ok {
		t.Fatalf("expected a standby relay")
	}
	if best.RelayID != "good" {
		t.Fatalf("standby = %s, want good (skip active, unstable and down)", best.RelayID)
	}

	// Only the active relay available: no standby.
	if _, ok := selectStandbyResult([]ProbeResult{{RelayID: "active", Endpoint: "a:1", Reachable: true}}, "a:1"); ok {
		t.Fatalf("the active relay must never be chosen as standby")
	}
}

func TestFreshStandbyExpiry(t *testing.T) {
	engine := &Engine{}
	result := ProbeResult{RelayID: "s1", Endpoint: "s:1", Reachable: true}

	if _, ok := engine.freshStandby(); ok {
		t.Fatalf("no standby must not be reported as fresh")
	}

	engine.standby = &result
	engine.standbyAt = time.Now()
	if got, ok := engine.freshStandby(); !ok || got.RelayID != "s1" {
		t.Fatalf("recent standby must be fresh: %+v ok=%v", got, ok)
	}

	engine.standbyAt = time.Now().Add(-2 * standbyMaxAge)
	if _, ok := engine.freshStandby(); ok {
		t.Fatalf("stale standby must not be fresh")
	}

	engine.clearStandby()
	if _, ok := engine.freshStandby(); ok {
		t.Fatalf("cleared standby must not be fresh")
	}
}
