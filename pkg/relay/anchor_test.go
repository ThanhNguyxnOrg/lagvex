package relay

import (
	"net"
	"testing"
	"time"
)

func TestAllowAnchorProbeRateLimit(t *testing.T) {
	sess := &Session{}
	now := time.Now()

	if !sess.AllowAnchorProbe(now) {
		t.Fatalf("first probe must be allowed")
	}
	if sess.AllowAnchorProbe(now.Add(500 * time.Millisecond)) {
		t.Fatalf("probe inside the minimum interval must be rejected")
	}
	if !sess.AllowAnchorProbe(now.Add(anchorProbeMinInterval + time.Millisecond)) {
		t.Fatalf("probe after the minimum interval must be allowed")
	}

	// Two probes were already counted above; exhaust the rest of the hourly
	// budget and confirm the window resets afterwards.
	base := now.Add(time.Minute)
	for i := 0; i < anchorProbeMaxPerHour-2; i++ {
		base = base.Add(anchorProbeMinInterval + time.Millisecond)
		if !sess.AllowAnchorProbe(base) {
			t.Fatalf("probe %d inside the hourly budget was rejected", i)
		}
	}
	if sess.AllowAnchorProbe(base.Add(anchorProbeMinInterval + time.Millisecond)) {
		t.Fatalf("probe over the hourly budget must be rejected")
	}
	if !sess.AllowAnchorProbe(base.Add(time.Hour + time.Minute)) {
		t.Fatalf("probe after the hour window must be allowed again")
	}
}

func TestProbeAnchorTargetMeasuresLocalListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	rttUs, lossPct := probeAnchorTarget(listener.Addr().String())
	if rttUs == 0 {
		t.Fatalf("expected a positive RTT to the local listener")
	}
	if lossPct != 0 {
		t.Fatalf("loss = %d%%, want 0", lossPct)
	}
}

func TestProbeAnchorTargetCountsFailures(t *testing.T) {
	// 127.0.0.1:1 is not listening; both samples fail.
	rttUs, lossPct := probeAnchorTarget("127.0.0.1:1")
	if rttUs != 0 {
		t.Fatalf("expected no RTT, got %d", rttUs)
	}
	if lossPct != 100 {
		t.Fatalf("loss = %d%%, want 100", lossPct)
	}
}
