package client

import (
	"bytes"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

func newTestDualSender() (*dualSender, *time.Time) {
	now := time.Unix(1_700_000_300, 0)
	sender := newDualSender(0xFEEDC0DE, func() time.Time { return now }, 40*time.Millisecond)
	return sender, &now
}

func TestDualSenderFramesSequentialTrailer(t *testing.T) {
	sender, _ := newTestDualSender()
	inner := bytes.Repeat([]byte{0x45}, 28)

	first, seq1 := sender.frame(inner)
	if seq1 != 1 {
		t.Fatalf("first seq = %d, want 1", seq1)
	}
	got, gotSeq, gotHash, ok := protocol.SplitDualTrailer(first)
	if !ok || gotSeq != 1 || gotHash != 0xFEEDC0DE {
		t.Fatalf("trailer = seq %d hash %08x ok %v", gotSeq, gotHash, ok)
	}
	if !bytes.Equal(got, inner) {
		t.Fatal("inner payload was modified by framing")
	}

	_, seq2 := sender.frame(inner)
	_, seq3 := sender.frame(inner)
	if seq2 != 2 || seq3 != 3 {
		t.Fatalf("seqs = %d,%d, want 2,3", seq2, seq3)
	}
}

func TestDualSenderAckPurgesContiguous(t *testing.T) {
	sender, _ := newTestDualSender()
	inner := bytes.Repeat([]byte{0x45}, 28)
	for i := 0; i < 8; i++ {
		sender.frame(inner)
	}

	sender.onAck(5)
	stats := sender.stats()
	if stats.InFlight != 3 {
		t.Fatalf("in flight = %d, want 3 (seqs 6..8 stay)", stats.InFlight)
	}
	if stats.AckPurged != 5 || stats.AckedSeq != 5 {
		t.Fatalf("purged=%d acked=%d, want 5/5", stats.AckPurged, stats.AckedSeq)
	}

	// A lower/older watermark never rolls the purge back.
	sender.onAck(3)
	if stats := sender.stats(); stats.AckedSeq != 5 {
		t.Fatalf("acked rolled back to %d", stats.AckedSeq)
	}
}

func TestDualSenderRetransmitAttempts(t *testing.T) {
	sender, now := newTestDualSender()
	inner := bytes.Repeat([]byte{0x45}, 28)
	sender.frame(inner)

	// Before the threshold: nothing due.
	if due := sender.pendingRetransmits(); len(due) != 0 {
		t.Fatalf("retransmits before threshold: %d", len(due))
	}

	// Attempt 1 (primary) after the threshold.
	*now = now.Add(41 * time.Millisecond)
	due := sender.pendingRetransmits()
	if len(due) != 1 || due[0].Seq != 1 || due[0].Attempt != 1 {
		t.Fatalf("due = %+v, want seq 1 attempt 1", due)
	}

	// Backoff: nothing due immediately after attempt 1.
	if due := sender.pendingRetransmits(); len(due) != 0 {
		t.Fatalf("immediate second attempt: %+v", due)
	}

	// Attempt 2 (standby) after the backoff.
	*now = now.Add(80 * time.Millisecond)
	due = sender.pendingRetransmits()
	if len(due) != 1 || due[0].Seq != 1 || due[0].Attempt != 2 {
		t.Fatalf("due = %+v, want seq 1 attempt 2", due)
	}

	// Attempts are capped.
	*now = now.Add(500 * time.Millisecond)
	if due := sender.pendingRetransmits(); len(due) != 0 {
		t.Fatalf("retransmits beyond cap: %+v", due)
	}
	if stats := sender.stats(); stats.Retransmits != 2 {
		t.Fatalf("retransmit counter = %d, want 2", stats.Retransmits)
	}
}

func TestDualSenderAckStopsRetransmit(t *testing.T) {
	sender, now := newTestDualSender()
	inner := bytes.Repeat([]byte{0x45}, 28)
	sender.frame(inner)

	*now = now.Add(41 * time.Millisecond)
	if due := sender.pendingRetransmits(); len(due) != 1 {
		t.Fatal("expected one retransmission")
	}

	// The ACK arrives late but before attempt 2: no standby copy.
	sender.onAck(1)
	if due := sender.pendingRetransmits(); len(due) != 0 {
		t.Fatalf("acked frame still due: %+v", due)
	}
	if stats := sender.stats(); stats.InFlight != 0 {
		t.Fatalf("in flight after ack = %d, want 0", stats.InFlight)
	}
}

func TestDualSenderRingOverflowAndExpiry(t *testing.T) {
	sender, now := newTestDualSender()
	inner := bytes.Repeat([]byte{0x45}, 28)

	for i := 0; i < dualRingCap+10; i++ {
		sender.frame(inner)
	}
	stats := sender.stats()
	if stats.InFlight != dualRingCap {
		t.Fatalf("in flight = %d, want cap %d", stats.InFlight, dualRingCap)
	}
	if stats.RingOverflow == 0 {
		t.Fatal("overflow counter must increase when the oldest frames are evicted unacked")
	}

	// Everything ages out after the TTL.
	*now = now.Add(2 * dualFrameTTL)
	if due := sender.pendingRetransmits(); len(due) != 0 {
		t.Fatalf("expired frames still due: %d", len(due))
	}
	if stats := sender.stats(); stats.InFlight != 0 || stats.Expired == 0 {
		t.Fatalf("in flight=%d expired=%d, want 0/nonzero", stats.InFlight, stats.Expired)
	}
}

func TestDualSenderThresholdClamped(t *testing.T) {
	fast := newDualSender(1, nil, time.Millisecond)
	if fast.threshold != DualMinRetransmitThreshold {
		t.Fatalf("threshold = %s, want >= min", fast.threshold)
	}
	slow := newDualSender(1, nil, time.Hour)
	if slow.threshold != DualMaxRetransmitThreshold {
		t.Fatalf("threshold = %s, want <= max", slow.threshold)
	}
}
