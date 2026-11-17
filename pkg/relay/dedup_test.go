package relay

import (
	"testing"
	"time"
)

func newTestRegistry() (*dualDedupRegistry, *time.Time) {
	now := time.Unix(1_700_000_000, 0)
	reg := &dualDedupRegistry{
		groups: make(map[uint32]*dualGroup),
		now:    func() time.Time { return now },
	}
	return reg, &now
}

func TestDualDedupFirstCopyWins(t *testing.T) {
	reg, _ := newTestRegistry()
	reg.touch(0x11111111)

	if reg.observe(0x11111111, 1) {
		t.Fatal("first observation must not be a duplicate")
	}
	if !reg.observe(0x11111111, 1) {
		t.Fatal("second copy of the same sequence must be a duplicate")
	}
	if reg.observe(0x11111111, 2) {
		t.Fatal("a different sequence must not be a duplicate")
	}
	// An unregistered group is never deduped: only PSK handshakes create groups.
	if reg.observe(0x22222222, 1) {
		t.Fatal("unregistered groups must not dedup")
	}
}

func TestDualDedupTTLSuppression(t *testing.T) {
	reg, now := newTestRegistry()
	reg.touch(0x33333333)

	if reg.observe(0x33333333, 9) {
		t.Fatal("first observation must not be a duplicate")
	}
	*now = now.Add(dualSeqTTL / 2)
	if !reg.observe(0x33333333, 9) {
		t.Fatal("within the TTL the copy must still be suppressed")
	}
	*now = now.Add(dualSeqTTL + time.Second)
	if reg.observe(0x33333333, 9) {
		t.Fatal("after the TTL the sequence is forgettable again")
	}
}

func TestDualDedupWindowCapEvictsOldest(t *testing.T) {
	reg, now := newTestRegistry()
	reg.touch(0x44444444)

	for seq := uint32(0); seq < dualSeqCap; seq++ {
		*now = now.Add(time.Microsecond)
		if reg.observe(0x44444444, seq) {
			t.Fatalf("seq %d unexpectedly marked duplicate during fill", seq)
		}
	}

	// One more forwarded packet evicts the oldest entry (seq 0).
	*now = now.Add(time.Microsecond)
	if reg.observe(0x44444444, dualSeqCap) {
		t.Fatal("new sequence must be forwarded")
	}
	if reg.observe(0x44444444, 0) {
		t.Fatal("evicted oldest entry must be forgettable")
	}
	if !reg.observe(0x44444444, dualSeqCap-1) {
		t.Fatal("recent entry must still be remembered")
	}
}

func TestDualDedupGroupCapEvictsLRU(t *testing.T) {
	reg, now := newTestRegistry()

	// Advance the fake clock between touches so group ages are distinct.
	for i := uint32(0); i < dualMaxGroups; i++ {
		*now = now.Add(time.Millisecond)
		reg.touch(i)
	}
	// Touch group 0 last so it is the most recently used.
	*now = now.Add(time.Millisecond)
	reg.touch(0)

	*now = now.Add(time.Millisecond)
	reg.touch(0xDEAD) // forces an eviction of the least recently used group
	if len(reg.groups) > dualMaxGroups {
		t.Fatalf("group cap exceeded: %d", len(reg.groups))
	}
	// Group 1 was the least recently used and must be gone; observing it must
	// not create dedup state.
	if reg.observe(1, 42) {
		t.Fatal("evicted group must not dedup")
	}
	if _, ok := reg.groups[1]; ok {
		t.Fatal("observe must not resurrect an evicted group")
	}
	// Group 0 survived.
	if _, ok := reg.groups[0]; !ok {
		t.Fatal("recently used group must survive eviction")
	}
}

func TestDualDedupSweepExpiresGroups(t *testing.T) {
	reg, now := newTestRegistry()
	reg.touch(0x55555555)
	if reg.observe(0x55555555, 3) {
		t.Fatal("unexpected duplicate")
	}

	*now = now.Add(dualGroupTTL + time.Minute)
	reg.sweep()
	if len(reg.groups) != 0 {
		t.Fatalf("expired group survived sweep: %d", len(reg.groups))
	}
}

func TestDualAckContiguousWatermark(t *testing.T) {
	now := time.Unix(1_700_000_100, 0)
	sess := &Session{dualAckHoles: make(map[uint32]struct{})}

	if _, ok := sess.TakeDualAck(now); ok {
		t.Fatal("no ACK before the first forwarded dual frame")
	}

	// In-order frames advance the watermark directly.
	sess.TrackDualForwarded(1)
	highest, ok := sess.TakeDualAck(now)
	if !ok || highest != 1 {
		t.Fatalf("first ack = %d/%v, want 1/true", highest, ok)
	}

	// A gap holds the watermark: frame 3 arrives before frame 2. The ACK
	// still fires (rate limit allows) but only reports the contiguous part.
	sess.TrackDualForwarded(3)
	highest, ok = sess.TakeDualAck(now.Add(dualAckInterval + time.Millisecond))
	if !ok || highest != 1 {
		t.Fatalf("ack across gap = %d/%v, want 1/true", highest, ok)
	}

	// Fill the gap: watermark jumps to 3.
	sess.TrackDualForwarded(2)
	highest, ok = sess.TakeDualAck(now.Add(2*dualAckInterval + time.Millisecond))
	if !ok || highest != 3 {
		t.Fatalf("ack after fill = %d/%v, want 3/true", highest, ok)
	}

	// A retransmitted duplicate fills a hole too (proof of earlier forward).
	sess.TrackDualForwarded(6)
	sess.TrackDualForwarded(5)
	highest, _ = sess.TakeDualAck(now.Add(4 * dualAckInterval))
	if highest != 3 {
		t.Fatalf("ack = %d, want 3 (5 and 6 are holes)", highest)
	}
	sess.TrackDualForwarded(4)
	sess.TrackDualForwarded(5) // suppressed duplicate of the hole
	highest, _ = sess.TakeDualAck(now.Add(6 * dualAckInterval))
	if highest != 6 {
		t.Fatalf("ack = %d, want 6", highest)
	}

	// Rate limit still applies between ACKs.
	if _, ok := sess.TakeDualAck(now.Add(6*dualAckInterval + time.Millisecond)); ok {
		t.Fatal("ACK must be rate limited")
	}
}

func TestDualAckHolesBounded(t *testing.T) {
	now := time.Unix(1_700_000_200, 0)
	sess := &Session{dualAckHoles: make(map[uint32]struct{})}

	// Watermark stuck at a permanent gap; push more holes than the cap.
	sess.TrackDualForwarded(1)
	sess.TrackDualForwarded(dualAckHolesCap + 10)
	for i := 2; i <= dualAckHolesCap+9; i++ {
		sess.TrackDualForwarded(uint32(i))
	}
	if len(sess.dualAckHoles) > dualAckHolesCap {
		t.Fatalf("hole set grew to %d", len(sess.dualAckHoles))
	}

	// The evicted oldest hole re-enters tracking when its retransmission
	// arrives as a suppressed duplicate.
	sess.TrackDualForwarded(2)
	highest, _ := sess.TakeDualAck(now.Add(dualAckInterval + time.Millisecond))
	if highest < 2 {
		t.Fatalf("ack = %d, want >= 2 after hole re-entry", highest)
	}
}

func TestDualDedupStatsSnapshot(t *testing.T) {
	reg, _ := newTestRegistry()
	reg.touch(0x66666666)
	reg.observe(0x66666666, 1)
	reg.observe(0x66666666, 1)
	reg.observe(0x66666666, 2)

	stats := reg.stats()
	if len(stats) != 1 {
		t.Fatalf("stats groups = %d, want 1", len(stats))
	}
	if stats[0].Forwarded != 2 || stats[0].Suppressed != 1 {
		t.Fatalf("counters = forwarded %d suppressed %d", stats[0].Forwarded, stats[0].Suppressed)
	}
	if stats[0].Window != 2 {
		t.Fatalf("window = %d, want 2", stats[0].Window)
	}
}
