package protocol

import (
	"sync"
	"time"
)

// FEC block size constants.
const (
	FECBlockLight      = 10
	FECBlockMedium     = 6
	FECBlockAggressive = 4
)

// rsDecoderBlockCap bounds how many partially-received RS blocks are tracked
// concurrently in FECDecoder so loss bursts cannot grow memory unboundedly.
const rsDecoderBlockCap = 64

// RecoveredPacket represents one lost data packet reconstructed by FEC.
type RecoveredPacket struct {
	Seq  uint64
	Data []byte
}

// FECEncoderConfig tunes the systematic parity generator.
type FECEncoderConfig struct {
	DefaultBlockSize   int           // Number of data packets per parity block (K)
	DefaultParityCount int           // Number of parity shards per block (M: 1 = XOR, 2..6 = Reed-Solomon)
	FlushInterval      time.Duration // Maximum time window to hold parity before emitting
}

// DefaultFECEncoderConfig returns production defaults for fast-paced gaming.
func DefaultFECEncoderConfig() FECEncoderConfig {
	return FECEncoderConfig{
		DefaultBlockSize:   FECBlockMedium,
		DefaultParityCount: 1,
		FlushInterval:      15 * time.Millisecond,
	}
}

// FECEncoder accumulates consecutive outgoing datagrams and generates either
// XOR (M=1) or Cauchy Reed-Solomon (M>=2) parity packets.
type FECEncoder struct {
	mu            sync.Mutex
	blockSize     int
	parityCount   int
	flushInterval time.Duration
	baseSeq       uint64
	count         int
	lengths       []uint16
	parity        []byte
	shards        [][]byte
	lastFlush     time.Time
}

// NewFECEncoder initializes an encoder with the given configuration.
func NewFECEncoder(cfg FECEncoderConfig) *FECEncoder {
	bs := cfg.DefaultBlockSize
	if bs <= 1 {
		bs = 6
	}
	pc := cfg.DefaultParityCount
	if pc <= 0 {
		pc = 1
	}
	if pc > FECMaxParityShards {
		pc = FECMaxParityShards
	}
	fi := cfg.FlushInterval
	if fi <= 0 {
		fi = 15 * time.Millisecond
	}
	return &FECEncoder{
		blockSize:     bs,
		parityCount:   pc,
		flushInterval: fi,
		lengths:       make([]uint16, 0, bs),
		shards:        make([][]byte, 0, bs),
		lastFlush:     time.Now(),
	}
}

// SetBlockSize updates the adaptive group size K (e.g. 4 to 10), or 0 for standby.
func (e *FECEncoder) SetBlockSize(k int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if k == 0 || (k >= 2 && k <= 32) {
		if e.blockSize != k && k == 0 {
			e.clearLocked()
		}
		e.blockSize = k
	}
}

// SetProfile updates both data block size K (0 or 2..16) and parity shard count
// M (1..6). M=1 uses legacy XOR; M>=2 uses Cauchy Reed-Solomon burst recovery.
func (e *FECEncoder) SetProfile(k, m int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if k == 0 {
		e.blockSize = 0
		e.parityCount = 0
		e.clearLocked()
		return
	}
	if k < 2 {
		k = 2
	}
	if k > FECMaxDataShards && m > 1 {
		k = FECMaxDataShards
	}
	if m < 1 {
		m = 1
	}
	if m > FECMaxParityShards {
		m = FECMaxParityShards
	}
	if e.blockSize != k || e.parityCount != m {
		e.clearLocked()
		e.blockSize = k
		e.parityCount = m
	}
}

// BlockSize returns the currently active block size K.
func (e *FECEncoder) BlockSize() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.blockSize
}

// ParityCount returns the currently active parity shard count M.
func (e *FECEncoder) ParityCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.blockSize <= 0 {
		return 0
	}
	if e.parityCount <= 0 {
		return 1
	}
	return e.parityCount
}

// AddPacket processes an outgoing data packet and returns a single encoded
// TypeFEC payload if the block is complete (legacy single-parity helper).
func (e *FECEncoder) AddPacket(seq uint64, data []byte) ([]byte, bool) {
	pkts := e.AddPacketMulti(seq, data)
	if len(pkts) == 0 {
		return nil, false
	}
	return pkts[0], true
}

// AddPacketMulti processes an outgoing data packet and returns 0..M encoded
// TypeFEC parity payloads when a block completes or flushes on a sequence gap.
func (e *FECEncoder) AddPacketMulti(seq uint64, data []byte) [][]byte {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.blockSize <= 0 || len(data) == 0 {
		return nil
	}

	var emitted [][]byte
	if e.count > 0 && seq != e.baseSeq+uint64(e.count) {
		emitted = append(emitted, e.emitBlockLocked()...)
	}

	if e.count == 0 {
		e.baseSeq = seq
		e.lastFlush = time.Now()
		e.lengths = e.lengths[:0]
		e.parity = e.parity[:0]
		e.shards = e.shards[:0]
	}

	e.count++
	e.lengths = append(e.lengths, uint16(len(data)))

	if e.parityCount > 1 {
		cpy := make([]byte, len(data))
		copy(cpy, data)
		e.shards = append(e.shards, cpy)
	} else {
		if len(data) > len(e.parity) {
			oldLen := len(e.parity)
			newParity := make([]byte, len(data))
			copy(newParity, e.parity)
			e.parity = newParity
			for i := 0; i < oldLen; i++ {
				e.parity[i] ^= data[i]
			}
			copy(e.parity[oldLen:], data[oldLen:])
		} else {
			for i := 0; i < len(data); i++ {
				e.parity[i] ^= data[i]
			}
		}
	}

	if e.count >= e.blockSize {
		emitted = append(emitted, e.emitBlockLocked()...)
	}
	return emitted
}

// CheckFlush emits a single parity packet if the flush interval elapsed.
func (e *FECEncoder) CheckFlush(now time.Time) ([]byte, bool) {
	pkts := e.CheckFlushMulti(now)
	if len(pkts) == 0 {
		return nil, false
	}
	return pkts[0], true
}

// CheckFlushMulti emits all M parity packets for a pending block once the flush
// interval has elapsed.
func (e *FECEncoder) CheckFlushMulti(now time.Time) [][]byte {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.blockSize <= 0 {
		return nil
	}
	if e.count >= 2 && now.Sub(e.lastFlush) >= e.flushInterval {
		return e.emitBlockLocked()
	}
	return nil
}

// Reset clears any accumulated packets without emitting parity.
func (e *FECEncoder) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clearLocked()
}

func (e *FECEncoder) clearLocked() {
	e.count = 0
	e.lengths = e.lengths[:0]
	e.parity = e.parity[:0]
	e.shards = e.shards[:0]
	e.baseSeq = 0
}

func (e *FECEncoder) emitBlockLocked() [][]byte {
	if e.count == 0 {
		return nil
	}

	var out [][]byte
	if e.parityCount > 1 && len(e.shards) == e.count {
		parities := rsEncodeBlock(e.shards, e.parityCount)
		out = make([][]byte, 0, len(parities))
		for idx, p := range parities {
			out = append(out, EncodeFECPayloadRS(
				e.baseSeq,
				uint8(e.count),
				uint8(e.parityCount),
				uint8(idx),
				e.lengths,
				p,
			))
		}
	} else {
		out = [][]byte{EncodeFECPayload(e.baseSeq, uint8(e.count), e.lengths, e.parity)}
	}

	e.clearLocked()
	e.lastFlush = time.Now()
	return out
}

type rsBlockState struct {
	count       uint8
	parityCount uint8
	lengths     []uint16
	parities    map[uint8][]byte
}

// FECDecoder maintains a history of received packets and reconstructs dropped
// packets from both XOR (M=1) and Reed-Solomon (M>=2) parity datagrams.
type FECDecoder struct {
	mu           sync.Mutex
	bufferSize   int
	recent       map[uint64][]byte
	seqHistory   []uint64
	rsBlocks     map[uint64]*rsBlockState
	rsBlockOrder []uint64
}

// NewFECDecoder initializes a decoder maintaining up to bufferSize recent packets.
func NewFECDecoder(bufferSize int) *FECDecoder {
	if bufferSize <= 0 {
		bufferSize = 128
	}
	return &FECDecoder{
		bufferSize:   bufferSize,
		recent:       make(map[uint64][]byte, bufferSize),
		seqHistory:   make([]uint64, 0, bufferSize),
		rsBlocks:     make(map[uint64]*rsBlockState, 16),
		rsBlockOrder: make([]uint64, 0, 16),
	}
}

// RecordPacket stores an incoming data packet for potential parity reconstruction.
func (d *FECDecoder) RecordPacket(seq uint64, data []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.recordLocked(seq, data)
}

func (d *FECDecoder) recordLocked(seq uint64, data []byte) {
	if _, exists := d.recent[seq]; exists {
		return
	}

	cpy := make([]byte, len(data))
	copy(cpy, data)
	d.recent[seq] = cpy
	d.seqHistory = append(d.seqHistory, seq)

	if len(d.seqHistory) > d.bufferSize {
		pruneCount := len(d.seqHistory) - d.bufferSize
		for i := 0; i < pruneCount; i++ {
			delete(d.recent, d.seqHistory[i])
		}
		d.seqHistory = d.seqHistory[pruneCount:]
	}
}

// ProcessParity evaluates an incoming FEC parity datagram and returns the first
// reconstructed packet (legacy single-packet helper).
func (d *FECDecoder) ProcessParity(payload FECPayload) (uint64, []byte, bool) {
	recovered := d.ProcessParityMulti(payload)
	if len(recovered) == 0 {
		return 0, nil, false
	}
	return recovered[0].Seq, recovered[0].Data, true
}

// ProcessParityMulti evaluates an incoming XOR or Reed-Solomon parity datagram
// and returns all reconstructed packets (up to M burst losses simultaneously).
func (d *FECDecoder) ProcessParityMulti(payload FECPayload) []RecoveredPacket {
	d.mu.Lock()
	defer d.mu.Unlock()

	if payload.Count == 0 || len(payload.Lengths) != int(payload.Count) {
		return nil
	}

	// Fast path: legacy single-parity XOR (ParityCount <= 1).
	if payload.ParityCount <= 1 {
		var missingSeq uint64
		var missingIdx int
		missingCount := 0

		for i := 0; i < int(payload.Count); i++ {
			seq := payload.BaseSeq + uint64(i)
			if _, ok := d.recent[seq]; !ok {
				missingSeq = seq
				missingIdx = i
				missingCount++
			}
		}

		if missingCount != 1 {
			return nil
		}

		targetLen := int(payload.Lengths[missingIdx])
		if targetLen <= 0 || targetLen > len(payload.Parity) {
			return nil
		}

		reconstructed := make([]byte, targetLen)
		copy(reconstructed, payload.Parity[:targetLen])

		for i := 0; i < int(payload.Count); i++ {
			if i == missingIdx {
				continue
			}
			seq := payload.BaseSeq + uint64(i)
			other := d.recent[seq]
			limit := len(other)
			if limit > targetLen {
				limit = targetLen
			}
			for b := 0; b < limit; b++ {
				reconstructed[b] ^= other[b]
			}
		}

		d.recordLocked(missingSeq, reconstructed)
		return []RecoveredPacket{{Seq: missingSeq, Data: reconstructed}}
	}

	// Reed-Solomon multi-parity path (ParityCount >= 2).
	if payload.ParityIndex >= payload.ParityCount || int(payload.Count) > FECMaxDataShards {
		return nil
	}

	blk, exists := d.rsBlocks[payload.BaseSeq]
	if !exists || blk.count != payload.Count || blk.parityCount != payload.ParityCount {
		blk = &rsBlockState{
			count:       payload.Count,
			parityCount: payload.ParityCount,
			lengths:     append([]uint16(nil), payload.Lengths...),
			parities:    make(map[uint8][]byte, payload.ParityCount),
		}
		if !exists {
			d.rsBlockOrder = append(d.rsBlockOrder, payload.BaseSeq)
			for len(d.rsBlockOrder) > rsDecoderBlockCap {
				evict := d.rsBlockOrder[0]
				d.rsBlockOrder = d.rsBlockOrder[1:]
				delete(d.rsBlocks, evict)
			}
		}
		d.rsBlocks[payload.BaseSeq] = blk
	}

	if _, dup := blk.parities[payload.ParityIndex]; !dup {
		pCopy := make([]byte, len(payload.Parity))
		copy(pCopy, payload.Parity)
		blk.parities[payload.ParityIndex] = pCopy
	}

	present := make(map[int][]byte, payload.Count)
	var missingIndices []int
	var missingLens []int
	for i := 0; i < int(payload.Count); i++ {
		seq := payload.BaseSeq + uint64(i)
		if pkt, ok := d.recent[seq]; ok {
			present[i] = pkt
		} else {
			missingIndices = append(missingIndices, i)
			missingLens = append(missingLens, int(payload.Lengths[i]))
		}
	}

	if len(missingIndices) == 0 {
		delete(d.rsBlocks, payload.BaseSeq)
		return nil
	}
	if len(blk.parities) < len(missingIndices) {
		return nil // Wait for additional parity shards in this block.
	}

	// Select L = len(missingIndices) received parity shards in deterministic index order.
	need := len(missingIndices)
	pIndices := make([]int, 0, need)
	pData := make([][]byte, 0, need)
	for idx := uint8(0); idx < payload.ParityCount && len(pIndices) < need; idx++ {
		if shard, ok := blk.parities[idx]; ok {
			pIndices = append(pIndices, int(idx))
			pData = append(pData, shard)
		}
	}

	shards, ok := rsRecoverMissing(int(payload.Count), present, missingIndices, missingLens, pIndices, pData)
	if !ok {
		return nil
	}

	out := make([]RecoveredPacket, 0, len(shards))
	for i, shard := range shards {
		seq := payload.BaseSeq + uint64(missingIndices[i])
		d.recordLocked(seq, shard)
		out = append(out, RecoveredPacket{Seq: seq, Data: shard})
	}
	delete(d.rsBlocks, payload.BaseSeq)
	return out
}

// Reset clears any stored packets and pending RS blocks in the decoder history.
func (d *FECDecoder) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.recent = make(map[uint64][]byte, d.bufferSize)
	d.seqHistory = d.seqHistory[:0]
	d.rsBlocks = make(map[uint64]*rsBlockState, 16)
	d.rsBlockOrder = d.rsBlockOrder[:0]
}

// AdaptiveFECController monitors real-time packet loss and adapts both the FEC
// block size K and parity count M (XOR at low loss, Reed-Solomon at higher loss).
type AdaptiveFECController struct {
	mu           sync.Mutex
	enabled      bool
	currentRatio string
	blockSize    int
	parityCount  int
}

// NewAdaptiveFECController initializes the controller.
func NewAdaptiveFECController() *AdaptiveFECController {
	return &AdaptiveFECController{
		enabled:      true,
		currentRatio: "Medium (6:1)",
		blockSize:    6,
		parityCount:  1,
	}
}

// SetEnabled toggles FEC globally.
func (a *AdaptiveFECController) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = enabled
}

// IsEnabled returns whether FEC is enabled.
func (a *AdaptiveFECController) IsEnabled() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.enabled
}

// Status returns the current ratio descriptor and block size K.
func (a *AdaptiveFECController) Status() (ratio string, blockSize int, active bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentRatio, a.blockSize, a.enabled && a.blockSize > 0
}

// ParityCount returns the currently active parity shard count M.
func (a *AdaptiveFECController) ParityCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.enabled || a.blockSize <= 0 {
		return 0
	}
	return a.parityCount
}

// UpdateLoss adapts the block size based on measured loss percentage (legacy helper).
func (a *AdaptiveFECController) UpdateLoss(lossPct float64) (blockSize int, changed bool) {
	k, _, changed := a.UpdateLossProfile(lossPct)
	return k, changed
}

// UpdateLossProfile selects the optimal (K data, M parity) FEC profile based on
// measured packet loss, automatically escalating from XOR to multi-parity
// Reed-Solomon for 5–15% burst loss environments.
func (a *AdaptiveFECController) UpdateLossProfile(lossPct float64) (blockSize, parityCount int, changed bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	prevK, prevM := a.blockSize, a.parityCount
	if !a.enabled {
		a.currentRatio = "Disabled"
		a.blockSize = 0
		a.parityCount = 0
		return 0, 0, prevK != 0 || prevM != 0
	}

	switch {
	case lossPct < 0.5:
		a.currentRatio = "Standby (0%)"
		a.blockSize = 0
		a.parityCount = 0
	case lossPct < 3.0:
		a.currentRatio = "Light XOR (10:1)"
		a.blockSize = 10
		a.parityCount = 1
	case lossPct < 8.0:
		a.currentRatio = "Medium RS (6:2)"
		a.blockSize = 6
		a.parityCount = 2
	case lossPct < 12.0:
		a.currentRatio = "High RS (4:2)"
		a.blockSize = 4
		a.parityCount = 2
	default:
		a.currentRatio = "Extreme RS (4:3)"
		a.blockSize = 4
		a.parityCount = 3
	}

	return a.blockSize, a.parityCount, a.blockSize != prevK || a.parityCount != prevM
}
