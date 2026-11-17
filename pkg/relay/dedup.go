package relay

// Dual-path dedup (Redundant-lite Phase 1). A dual-mode client sends one data
// frame to two relays; the first relay to receive a (group, sequence) pair
// forwards it and every later copy is dropped. The registry is bounded on both
// axes: per-group sequence windows and the number of tracked groups.
//
// A group only exists here after a PSK-authenticated handshake registered it,
// and the group hash is derived from the PSK, so unauthenticated traffic can
// never create or poison dedup state.

import (
	"sync"
	"time"
)

const (
	// dualSeqTTL is how long a forwarded sequence is remembered. It must
	// exceed the plausible spread between the two relays' delivery times.
	dualSeqTTL = 2 * time.Second

	// dualSeqCap bounds one group's sequence window; the oldest entries are
	// evicted once the cap is reached.
	dualSeqCap = 4096

	// dualGroupTTL is how long a group without observations is kept.
	dualGroupTTL = 10 * time.Minute

	// dualMaxGroups bounds total dedup state.
	dualMaxGroups = 256
)

// dualGroup holds one connection group's dedup window and counters.
type dualGroup struct {
	window     map[uint32]time.Time
	lastSeen   time.Time
	forwarded  uint64
	suppressed uint64
}

// dualDedupRegistry tracks first-copy-wins state for dual-path groups.
// now is injectable so tests can exercise TTLs without sleeping.
type dualDedupRegistry struct {
	mu     sync.Mutex
	groups map[uint32]*dualGroup
	now    func() time.Time
}

func newDualDedupRegistry() *dualDedupRegistry {
	return &dualDedupRegistry{
		groups: make(map[uint32]*dualGroup),
		now:    time.Now,
	}
}

// touch registers or refreshes a group. Called on the dual handshake.
func (r *dualDedupRegistry) touch(groupHash uint32) {
	now := r.now()
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.groups[groupHash]
	if !ok {
		if len(r.groups) >= dualMaxGroups {
			r.evictGroupsLocked(now)
		}
		g = &dualGroup{window: make(map[uint32]time.Time), lastSeen: now}
		r.groups[groupHash] = g
	}
	g.lastSeen = now
}

// observe reports whether this (group, sequence) was already forwarded within
// the TTL. Unregistered groups are never deduped.
func (r *dualDedupRegistry) observe(groupHash, seq uint32) bool {
	now := r.now()
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.groups[groupHash]
	if !ok {
		return false
	}
	g.lastSeen = now

	if ts, seen := g.window[seq]; seen && now.Sub(ts) < dualSeqTTL {
		g.suppressed++
		return true
	}
	if len(g.window) >= dualSeqCap {
		r.evictWindowLocked(g, now)
	}
	g.window[seq] = now
	g.forwarded++
	return false
}

// evictWindowLocked drops expired sequences, then the oldest ones, until the
// window is below the cap. Callers hold r.mu.
func (r *dualDedupRegistry) evictWindowLocked(g *dualGroup, now time.Time) {
	for seq, ts := range g.window {
		if now.Sub(ts) >= dualSeqTTL {
			delete(g.window, seq)
		}
	}
	for len(g.window) >= dualSeqCap {
		oldestSeq, oldestTS, found := uint32(0), time.Time{}, false
		for seq, ts := range g.window {
			if !found || ts.Before(oldestTS) {
				oldestSeq, oldestTS, found = seq, ts, true
			}
		}
		if !found {
			return
		}
		delete(g.window, oldestSeq)
	}
}

// evictGroupsLocked drops expired groups, then the least recently used ones,
// until there is room for one more group. Callers hold r.mu.
func (r *dualDedupRegistry) evictGroupsLocked(now time.Time) {
	for hash, g := range r.groups {
		if now.Sub(g.lastSeen) >= dualGroupTTL {
			delete(r.groups, hash)
		}
	}
	for len(r.groups) >= dualMaxGroups {
		oldestHash, oldestTS, found := uint32(0), time.Time{}, false
		for hash, g := range r.groups {
			if !found || g.lastSeen.Before(oldestTS) {
				oldestHash, oldestTS, found = hash, g.lastSeen, true
			}
		}
		if !found {
			return
		}
		delete(r.groups, oldestHash)
	}
}

// sweep removes expired sequences and groups. The reaper calls it periodically.
func (r *dualDedupRegistry) sweep() {
	now := r.now()
	r.mu.Lock()
	defer r.mu.Unlock()

	for hash, g := range r.groups {
		if now.Sub(g.lastSeen) >= dualGroupTTL {
			delete(r.groups, hash)
			continue
		}
		for seq, ts := range g.window {
			if now.Sub(ts) >= dualSeqTTL {
				delete(g.window, seq)
			}
		}
	}
}

// DualDedupStats is a snapshot of one group's counters.
type DualDedupStats struct {
	GroupHash  uint32 `json:"groupHash"`
	Forwarded  uint64 `json:"forwarded"`
	Suppressed uint64 `json:"suppressed"`
	Window     int    `json:"window"`
}

// stats snapshots the live groups, most recently used first.
func (r *dualDedupRegistry) stats() []DualDedupStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]DualDedupStats, 0, len(r.groups))
	for hash, g := range r.groups {
		out = append(out, DualDedupStats{
			GroupHash:  hash,
			Forwarded:  g.forwarded,
			Suppressed: g.suppressed,
			Window:     len(g.window),
		})
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].GroupHash < out[j-1].GroupHash; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
