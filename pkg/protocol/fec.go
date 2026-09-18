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

// FECEncoderConfig tunes the systematic parity generator.
type FECEncoderConfig struct {
	DefaultBlockSize int           // Number of data packets per parity datagram (K)
	FlushInterval    time.Duration // Maximum time window to hold parity before emitting
}

// DefaultFECEncoderConfig returns production defaults for fast-paced gaming.
func DefaultFECEncoderConfig() FECEncoderConfig {
	return FECEncoderConfig{
		DefaultBlockSize: FECBlockMedium,
		FlushInterval:    15 * time.Millisecond,
	}
}

// FECEncoder accumulates consecutive outgoing datagrams and generates XOR parity packets.
type FECEncoder struct {
	mu            sync.Mutex
	blockSize     int
	flushInterval time.Duration
	baseSeq       uint64
	count         int
	lengths       []uint16
	parity        []byte
	lastFlush     time.Time
}

// NewFECEncoder initializes an encoder with the given configuration.
func NewFECEncoder(cfg FECEncoderConfig) *FECEncoder {
	bs := cfg.DefaultBlockSize
	if bs <= 1 {
		bs = 6
	}
	fi := cfg.FlushInterval
	if fi <= 0 {
		fi = 15 * time.Millisecond
	}
	return &FECEncoder{
		blockSize:     bs,
		flushInterval: fi,
		lengths:       make([]uint16, 0, bs),
		lastFlush:     time.Now(),
	}
}

// SetBlockSize updates the adaptive group size K (e.g. 4 to 10).
func (e *FECEncoder) SetBlockSize(k int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if k >= 2 && k <= 32 {
		e.blockSize = k
	}
}

// BlockSize returns the currently active block size K.
func (e *FECEncoder) BlockSize() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.blockSize
}

// AddPacket processes an outgoing data packet and returns an encoded TypeFEC payload
// if the block is complete.
func (e *FECEncoder) AddPacket(seq uint64, data []byte) ([]byte, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(data) == 0 {
		return nil, false
	}

	if e.count == 0 {
		e.baseSeq = seq
		e.lastFlush = time.Now()
		e.lengths = e.lengths[:0]
		e.parity = e.parity[:0]
	}

	e.count++
	e.lengths = append(e.lengths, uint16(len(data)))

	// Extend parity slice if this packet is longer than current max
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

	// Emit parity if group size reached
	if e.count >= e.blockSize {
		return e.emitBlockLocked()
	}

	return nil, false
}

// CheckFlush emits a parity packet if the current group has pending packets and
// the flush interval has elapsed.
func (e *FECEncoder) CheckFlush(now time.Time) ([]byte, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.count >= 2 && now.Sub(e.lastFlush) >= e.flushInterval {
		return e.emitBlockLocked()
	}
	return nil, false
}

// Reset clears any accumulated packets without emitting parity.
func (e *FECEncoder) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.count = 0
	e.lengths = e.lengths[:0]
	e.parity = e.parity[:0]
	e.baseSeq = 0
}

func (e *FECEncoder) emitBlockLocked() ([]byte, bool) {
	if e.count == 0 {
		return nil, false
	}
	payload := EncodeFECPayload(e.baseSeq, uint8(e.count), e.lengths, e.parity)

	e.count = 0
	e.lengths = e.lengths[:0]
	e.parity = e.parity[:0]
	e.baseSeq = 0
	e.lastFlush = time.Now()

	return payload, true
}

// FECDecoder maintains a history of received packets and reconstructs dropped packets
// upon receiving TypeFEC parity datagrams.
type FECDecoder struct {
	mu         sync.Mutex
	bufferSize int
	recent     map[uint64][]byte
	seqHistory []uint64
}

// NewFECDecoder initializes a decoder maintaining up to bufferSize recent packets.
func NewFECDecoder(bufferSize int) *FECDecoder {
	if bufferSize <= 0 {
		bufferSize = 128
	}
	return &FECDecoder{
		bufferSize: bufferSize,
		recent:     make(map[uint64][]byte, bufferSize),
		seqHistory: make([]uint64, 0, bufferSize),
	}
}

// RecordPacket stores an incoming data packet for potential parity reconstruction.
func (d *FECDecoder) RecordPacket(seq uint64, data []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

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

// ProcessParity evaluates an incoming FEC parity datagram and attempts single-loss recovery.
// Returns (reconstructedSeq, reconstructedData, true) if a dropped packet was restored.
func (d *FECDecoder) ProcessParity(payload FECPayload) (uint64, []byte, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if payload.Count == 0 || len(payload.Lengths) != int(payload.Count) {
		return 0, nil, false
	}

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
		return 0, nil, false
	}

	targetLen := int(payload.Lengths[missingIdx])
	if targetLen <= 0 || targetLen > len(payload.Parity) {
		return 0, nil, false
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

	d.recent[missingSeq] = reconstructed
	d.seqHistory = append(d.seqHistory, missingSeq)

	return missingSeq, reconstructed, true
}

// Reset clears any stored packets in the decoder history.
func (d *FECDecoder) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.recent = make(map[uint64][]byte, d.bufferSize)
	d.seqHistory = d.seqHistory[:0]
}

// AdaptiveFECController monitors real-time packet loss and adapts the FEC redundancy ratio.
type AdaptiveFECController struct {
	mu           sync.Mutex
	enabled      bool
	currentRatio string
	blockSize    int
}

// NewAdaptiveFECController initializes the controller.
func NewAdaptiveFECController() *AdaptiveFECController {
	return &AdaptiveFECController{
		enabled:      true,
		currentRatio: "Off",
		blockSize:    0,
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

// UpdateLoss adapts the block size based on measured loss percentage.
func (a *AdaptiveFECController) UpdateLoss(lossPct float64) (blockSize int, active bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.enabled {
		a.currentRatio = "Disabled"
		a.blockSize = 0
		return 0, false
	}

	switch {
	case lossPct < 0.5:
		a.currentRatio = "Standby (0%)"
		a.blockSize = 0
		return 0, false
	case lossPct < 3.0:
		a.currentRatio = "Light (10:1)"
		a.blockSize = 10
		return 10, true
	case lossPct < 8.0:
		a.currentRatio = "Medium (6:1)"
		a.blockSize = 6
		return 6, true
	default:
		a.currentRatio = "Aggressive (4:1)"
		a.blockSize = 4
		return 4, true
	}
}
