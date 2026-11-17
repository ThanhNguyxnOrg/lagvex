package client

// Client-side dual-path sender (Redundant-lite Phase 2). Data frames go to
// the primary relay only; frames the primary has not acknowledged within a
// measured threshold are retransmitted — first to the primary relay (Phase 1
// relay dedup makes that idempotent), then to the standby relay. This keeps
// the common case free of duplicate packets at the game server.
//
// Everything here is bounded: the ring holds at most dualRingCap frames,
// each frame is retransmitted at most dualMaxRetransmits times, and entries
// expire after dualFrameTTL so a dead ACK stream cannot grow memory.

import (
	"sync"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

const (
	// dualRingCap bounds how many recent dual frames are eligible for
	// retransmission (~360 ms of game traffic at 1400 B and 500 pps).
	dualRingCap = 256

	// dualMaxRetransmits caps attempts per frame: one to the primary relay,
	// one to the standby relay.
	dualMaxRetransmits = 2

	// dualFrameTTL is how long a frame stays eligible regardless of ACKs.
	dualFrameTTL = time.Second

	// DualMinRetransmitThreshold and DualMaxRetransmitThreshold bound the
	// caller-provided threshold (typically max(40 ms, 1.5× RTT)).
	DualMinRetransmitThreshold = 20 * time.Millisecond
	DualMaxRetransmitThreshold = 250 * time.Millisecond
)

// DualRetransmit is one frame due for retransmission.
type DualRetransmit struct {
	Seq uint32
	// Frame is the sealed trailer payload (inner packet + dual trailer),
	// ready to be encrypted toward either relay.
	Frame []byte
	// Attempt counts retransmissions already spent; 1 → primary, 2 → standby.
	Attempt int
}

// DualSenderStats is a measured snapshot of the sender's state.
type DualSenderStats struct {
	GroupHash    uint32 `json:"groupHash"`
	AckedSeq     uint32 `json:"ackedSeq"`
	InFlight     int    `json:"inFlight"`
	Framed       uint64 `json:"framed"`
	Retransmits  uint64 `json:"retransmits"`
	AckPurged    uint64 `json:"ackPurged"`
	Expired      uint64 `json:"expired"`
	RingOverflow uint64 `json:"ringOverflow"`
	ThresholdMs  int    `json:"thresholdMs"`
}

type dualFrame struct {
	seq      uint32
	framed   []byte
	sentAt   time.Time
	attempts int
	nextTry  time.Time
}

type dualSender struct {
	mu        sync.Mutex
	groupHash uint32
	nextSeq   uint32
	ackedSeq  uint32
	ring      map[uint32]*dualFrame
	ringQ     []uint32

	now       func() time.Time
	threshold time.Duration

	framed       uint64
	retransmits  uint64
	ackPurged    uint64
	expired      uint64
	ringOverflow uint64
}

func newDualSender(groupHash uint32, now func() time.Time, threshold time.Duration) *dualSender {
	if now == nil {
		now = time.Now
	}
	if threshold < DualMinRetransmitThreshold {
		threshold = DualMinRetransmitThreshold
	}
	if threshold > DualMaxRetransmitThreshold {
		threshold = DualMaxRetransmitThreshold
	}
	return &dualSender{
		groupHash: groupHash,
		ring:      make(map[uint32]*dualFrame),
		now:       now,
		threshold: threshold,
	}
}

// frame assigns the next dual sequence, appends the trailer and stores the
// frame for potential retransmission. The returned slice is a fresh copy the
// caller may seal toward any relay.
func (d *dualSender) frame(inner []byte) ([]byte, uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.nextSeq++ // sequence 0 is reserved and never assigned
	seq := d.nextSeq
	framed := protocol.AppendDualTrailer(inner, seq, d.groupHash)

	d.ring[seq] = &dualFrame{
		seq:     seq,
		framed:  append([]byte(nil), framed...),
		sentAt:  now,
		nextTry: now.Add(d.threshold),
	}
	d.ringQ = append(d.ringQ, seq)
	d.framed++

	for len(d.ring) > dualRingCap {
		oldest := d.ringQ[0]
		d.ringQ = d.ringQ[1:]
		if entry, ok := d.ring[oldest]; ok {
			if entry.seq > d.ackedSeq {
				d.ringOverflow++
			}
			delete(d.ring, oldest)
		}
	}
	return framed, seq
}

// onAck applies a contiguous ACK from the primary relay: every sequence at or
// below the watermark was forwarded to the game server and needs no
// retransmission. Contiguity is what makes the purge safe.
func (d *dualSender) onAck(contiguous uint32) {
	if contiguous == 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if contiguous <= d.ackedSeq {
		return
	}
	d.ackedSeq = contiguous
	for seq, entry := range d.ring {
		if seq <= contiguous {
			delete(d.ring, seq)
			d.ackPurged++
		} else {
			_ = entry
		}
	}
	d.compactQueueLocked()
}

// pendingRetransmits returns frames whose ACK never arrived within the
// threshold, marking them so each attempt is sent at most once. Callers route
// by Attempt: 1 → primary relay, 2 → standby relay.
func (d *dualSender) pendingRetransmits() []DualRetransmit {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.expireLocked(now)

	var due []DualRetransmit
	for _, seq := range d.ringQ {
		entry, ok := d.ring[seq]
		if !ok {
			continue
		}
		if entry.seq <= d.ackedSeq || entry.attempts >= dualMaxRetransmits {
			continue
		}
		if now.Before(entry.nextTry) {
			continue
		}
		entry.attempts++
		d.retransmits++
		backoff := d.threshold * time.Duration(1<<entry.attempts)
		if backoff > dualFrameTTL/2 {
			backoff = dualFrameTTL / 2
		}
		entry.nextTry = now.Add(backoff)
		due = append(due, DualRetransmit{
			Seq:     entry.seq,
			Frame:   append([]byte(nil), entry.framed...),
			Attempt: entry.attempts,
		})
		if len(due) >= dualRingCap {
			break
		}
	}
	return due
}

// expireLocked drops frames older than dualFrameTTL: either their ACK stream
// died (failover territory) or they were delivered without our knowledge.
func (d *dualSender) expireLocked(now time.Time) {
	for _, seq := range d.ringQ {
		entry, ok := d.ring[seq]
		if !ok {
			continue
		}
		if now.Sub(entry.sentAt) >= dualFrameTTL {
			delete(d.ring, seq)
			d.expired++
		}
	}
	d.compactQueueLocked()
}

// compactQueueLocked rebuilds the FIFO queue from live ring entries. Callers
// hold d.mu.
func (d *dualSender) compactQueueLocked() {
	if len(d.ringQ) == len(d.ring) {
		return
	}
	live := d.ringQ[:0:0]
	for _, seq := range d.ringQ {
		if _, ok := d.ring[seq]; ok {
			live = append(live, seq)
		}
	}
	d.ringQ = live
}

// stats snapshots the measured counters.
func (d *dualSender) stats() DualSenderStats {
	d.mu.Lock()
	defer d.mu.Unlock()
	return DualSenderStats{
		GroupHash:    d.groupHash,
		AckedSeq:     d.ackedSeq,
		InFlight:     len(d.ring),
		Framed:       d.framed,
		Retransmits:  d.retransmits,
		AckPurged:    d.ackPurged,
		Expired:      d.expired,
		RingOverflow: d.ringOverflow,
		ThresholdMs:  int(d.threshold / time.Millisecond),
	}
}
