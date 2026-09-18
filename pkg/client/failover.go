package client

import (
	"fmt"
	"sync"
	"time"
)

// FailoverConfig defines parameters for hysteresis-based automatic relay switching.
type FailoverConfig struct {
	// MinScoreImprovement is the minimum percentage improvement required to switch (e.g. 0.15 = 15%).
	MinScoreImprovement float64
	// CooldownDuration is the minimum time between consecutive failovers to prevent route flapping.
	CooldownDuration time.Duration
	// MissedPongsThreshold is the consecutive missed keepalive pongs that trigger degradation.
	MissedPongsThreshold int
	// RTTSpikeFactor is the ratio of current RTT over baseline RTT that constitutes a spike (e.g. 1.4 = +40%).
	RTTSpikeFactor float64
	// PacketLossThreshold is the loss percentage triggering a candidate sweep (e.g. 3.0 = 3%).
	PacketLossThreshold float64
	// MaxHistory holds the maximum number of failover events recorded in memory.
	MaxHistory int
}

// DefaultFailoverConfig returns production-tuned defaults.
func DefaultFailoverConfig() FailoverConfig {
	return FailoverConfig{
		MinScoreImprovement:  0.15, // 15% better composite score required
		CooldownDuration:     5 * time.Second,
		MissedPongsThreshold: 3,
		RTTSpikeFactor:       1.40, // 40% latency surge
		PacketLossThreshold:  3.0,  // 3% loss
		MaxHistory:           20,
	}
}

// FailoverEvent logs a dynamic relay handover.
type FailoverEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	OldRelayID   string    `json:"oldRelayId"`
	OldRelayName string    `json:"oldRelayName"`
	NewRelayID   string    `json:"newRelayId"`
	NewRelayName string    `json:"newRelayName"`
	OldScore     float64   `json:"oldScore"`
	NewScore     float64   `json:"newScore"`
	OldPingMs    float64   `json:"oldPingMs"`
	NewPingMs    float64   `json:"newPingMs"`
	Reason       string    `json:"reason"`
	HandoverOk   bool      `json:"handoverOk"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// FailoverController manages the health evaluation, hysteresis math, and history of relay handovers.
type FailoverController struct {
	cfg        FailoverConfig
	mu         sync.Mutex
	lastSwitch time.Time
	history    []FailoverEvent
}

// NewFailoverController initializes a new controller.
func NewFailoverController(cfg FailoverConfig) *FailoverController {
	if cfg.MinScoreImprovement <= 0 {
		cfg.MinScoreImprovement = 0.15
	}
	if cfg.CooldownDuration <= 0 {
		cfg.CooldownDuration = 5 * time.Second
	}
	if cfg.MaxHistory <= 0 {
		cfg.MaxHistory = 20
	}
	return &FailoverController{
		cfg:     cfg,
		history: make([]FailoverEvent, 0, cfg.MaxHistory),
	}
}

// IsDegraded evaluates whether the current active relay is experiencing network degradation.
func (fc *FailoverController) IsDegraded(consecutiveMissedPongs int, currentRTT, baselineRTT, currentLoss float64) (bool, string) {
	if consecutiveMissedPongs >= fc.cfg.MissedPongsThreshold {
		return true, fmt.Sprintf("missed %d consecutive keepalive pongs", consecutiveMissedPongs)
	}

	if currentLoss >= fc.cfg.PacketLossThreshold {
		return true, fmt.Sprintf("packet loss surge: %.1f%% (threshold: %.1f%%)", currentLoss, fc.cfg.PacketLossThreshold)
	}

	if baselineRTT > 0 && currentRTT > 0 {
		spikeThreshold := baselineRTT * fc.cfg.RTTSpikeFactor
		if currentRTT >= spikeThreshold && (currentRTT-baselineRTT) >= 15.0 {
			return true, fmt.Sprintf("latency surge: %.1fms vs baseline %.1fms (+%.1f%%)",
				currentRTT, baselineRTT, ((currentRTT-baselineRTT)/baselineRTT)*100)
		}
	}

	return false, ""
}

// ShouldSwitch determines if a candidate relay satisfies the anti-flapping hysteresis requirement
// and cooldown period. Lower score is better.
func (fc *FailoverController) ShouldSwitch(currentScore, candidateScore float64, now time.Time) (bool, string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// 1. Check Cooldown
	if !fc.lastSwitch.IsZero() {
		elapsed := now.Sub(fc.lastSwitch)
		if elapsed < fc.cfg.CooldownDuration {
			return false, fmt.Sprintf("in cooldown: %v remaining", fc.cfg.CooldownDuration-elapsed)
		}
	}

	// 2. If current node is completely dead (score is infinite or <= 0 invalid), switch immediately
	if currentScore <= 0 || currentScore > 999999 {
		return true, "current relay unreachable"
	}

	// 3. Hysteresis check: candidate must be at least (1 - MinScoreImprovement) of current score
	// E.g. with 15% improvement, candidateScore must be <= currentScore * 0.85
	targetThreshold := currentScore * (1.0 - fc.cfg.MinScoreImprovement)
	if candidateScore <= targetThreshold {
		improvementPct := ((currentScore - candidateScore) / currentScore) * 100.0
		return true, fmt.Sprintf("qualifies: %.1f%% improvement (candidate score %.2f vs current %.2f)",
			improvementPct, candidateScore, currentScore)
	}

	return false, fmt.Sprintf("insufficient improvement: candidate score %.2f does not beat threshold %.2f",
		candidateScore, targetThreshold)
}

// RecordSwitch records a handover event and resets the cooldown timer.
func (fc *FailoverController) RecordSwitch(event FailoverEvent) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if event.HandoverOk {
		fc.lastSwitch = event.Timestamp
	}

	fc.history = append(fc.history, event)
	if len(fc.history) > fc.cfg.MaxHistory {
		fc.history = fc.history[len(fc.history)-fc.cfg.MaxHistory:]
	}
}

// History returns a copy of recent handover events.
func (fc *FailoverController) History() []FailoverEvent {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	result := make([]FailoverEvent, len(fc.history))
	copy(result, fc.history)
	return result
}

// LastSwitchTime returns when the last handover occurred.
func (fc *FailoverController) LastSwitchTime() time.Time {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.lastSwitch
}
