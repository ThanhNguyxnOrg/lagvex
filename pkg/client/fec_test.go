package client

import (
	"bytes"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

func TestFECExactRecoverySingleLoss(t *testing.T) {
	encoder := NewFECEncoder(FECEncoderConfig{
		DefaultBlockSize: 4,
		FlushInterval:    100 * time.Millisecond,
	})

	// 4 data packets of varying sizes (simulating game packets)
	p0 := []byte{0x45, 0x00, 0x00, 0x34, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	p1 := []byte{0x45, 0x00, 0x00, 0x40, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78}
	p2 := []byte{0x45, 0x00, 0x00, 0x28, 0x99, 0x88, 0x77, 0x66}
	p3 := []byte{0x45, 0x00, 0x00, 0x50, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c}

	packets := [][]byte{p0, p1, p2, p3}

	var parityPayload []byte
	var ready bool

	for i, p := range packets {
		seq := uint64(100 + i)
		parityPayload, ready = encoder.AddPacket(seq, p)
		if i < 3 && ready {
			t.Fatalf("expected not ready at packet %d", i)
		}
	}

	if !ready || len(parityPayload) == 0 {
		t.Fatalf("expected parity payload to be ready after 4 packets")
	}

	fecPayload, err := protocol.DecodeFECPayload(parityPayload)
	if err != nil {
		t.Fatalf("DecodeFECPayload failed: %v", err)
	}

	if fecPayload.BaseSeq != 100 || fecPayload.Count != 4 {
		t.Fatalf("unexpected BaseSeq %d or Count %d", fecPayload.BaseSeq, fecPayload.Count)
	}

	// Test dropping packet 1 (p1, seq 101)
	decoder := NewFECDecoder(64)
	decoder.RecordPacket(100, p0)
	// Skip 101 (dropped!)
	decoder.RecordPacket(102, p2)
	decoder.RecordPacket(103, p3)

	recSeq, recData, recovered := decoder.ProcessParity(fecPayload)
	if !recovered {
		t.Fatalf("expected packet 101 to be recovered, got false")
	}

	if recSeq != 101 {
		t.Errorf("expected recovered seq 101, got %d", recSeq)
	}

	if !bytes.Equal(recData, p1) {
		t.Fatalf("reconstructed data does not match original p1!\nOriginal:      %x\nReconstructed: %x", p1, recData)
	}
}

func TestFECRecoveryAllPacketsInTurn(t *testing.T) {
	encoder := NewFECEncoder(FECEncoderConfig{
		DefaultBlockSize: 3,
		FlushInterval:    50 * time.Millisecond,
	})

	origPackets := [][]byte{
		[]byte("competitive-gunfire-tick-packet-1"),
		[]byte("short-ping-ack"),
		[]byte("player-coordinates-vector-x-y-z-pitch-yaw-roll-1234567890"),
	}

	var parityPayload []byte
	var ready bool
	for i, p := range origPackets {
		parityPayload, ready = encoder.AddPacket(uint64(200+i), p)
	}
	if !ready {
		t.Fatalf("expected parity ready")
	}

	fecPayload, _ := protocol.DecodeFECPayload(parityPayload)

	// Test dropping each packet in turn (0, 1, 2)
	for dropIdx := 0; dropIdx < len(origPackets); dropIdx++ {
		decoder := NewFECDecoder(32)
		for i, p := range origPackets {
			if i == dropIdx {
				continue // drop!
			}
			decoder.RecordPacket(uint64(200+i), p)
		}

		expectedSeq := uint64(200 + dropIdx)
		expectedData := origPackets[dropIdx]

		recSeq, recData, recovered := decoder.ProcessParity(fecPayload)
		if !recovered {
			t.Fatalf("failed to recover dropped packet at index %d", dropIdx)
		}
		if recSeq != expectedSeq {
			t.Errorf("dropIdx %d: expected seq %d, got %d", dropIdx, expectedSeq, recSeq)
		}
		if !bytes.Equal(recData, expectedData) {
			t.Errorf("dropIdx %d: reconstructed data mismatch", dropIdx)
		}
	}
}

func TestFECZeroLossNoAction(t *testing.T) {
	encoder := NewFECEncoder(DefaultFECEncoderConfig())
	p0 := []byte("packet-0")
	p1 := []byte("packet-1")
	p2 := []byte("packet-2")

	encoder.AddPacket(1, p0)
	encoder.AddPacket(2, p1)
	encoder.AddPacket(3, p2)
	payload, ready := encoder.CheckFlush(time.Now().Add(20 * time.Millisecond))
	if !ready {
		t.Fatalf("expected flush ready")
	}

	fecPayload, _ := protocol.DecodeFECPayload(payload)

	decoder := NewFECDecoder(32)
	decoder.RecordPacket(1, p0)
	decoder.RecordPacket(2, p1)
	decoder.RecordPacket(3, p2)

	// Zero packets lost -> should return recovered = false
	_, _, recovered := decoder.ProcessParity(fecPayload)
	if recovered {
		t.Fatalf("expected no recovery when zero packets lost")
	}
}

func TestFECMultiLossNoFalseRecovery(t *testing.T) {
	encoder := NewFECEncoder(FECEncoderConfig{DefaultBlockSize: 4})
	p0 := []byte("packet-0")
	p1 := []byte("packet-1")
	p2 := []byte("packet-2")
	p3 := []byte("packet-3")

	encoder.AddPacket(10, p0)
	encoder.AddPacket(11, p1)
	encoder.AddPacket(12, p2)
	payload, _ := encoder.AddPacket(13, p3)

	fecPayload, _ := protocol.DecodeFECPayload(payload)

	decoder := NewFECDecoder(32)
	// Drop 2 packets: 11 and 12!
	decoder.RecordPacket(10, p0)
	decoder.RecordPacket(13, p3)

	_, _, recovered := decoder.ProcessParity(fecPayload)
	if recovered {
		t.Fatalf("expected false when multiple packets are missing in same block")
	}
}

func TestAdaptiveFECControllerScaling(t *testing.T) {
	ctrl := NewAdaptiveFECController()

	// Default state
	if !ctrl.IsEnabled() {
		t.Errorf("expected enabled by default")
	}

	// 0.2% loss -> Off / standby
	bs, active := ctrl.UpdateLoss(0.2)
	if active || bs != 0 {
		t.Errorf("expected standby at 0.2%% loss, got bs=%d active=%v", bs, active)
	}

	// 1.5% loss -> Light (10:1)
	bs, active = ctrl.UpdateLoss(1.5)
	if !active || bs != 10 {
		t.Errorf("expected Light (10:1) at 1.5%% loss, got bs=%d active=%v", bs, active)
	}

	// 5.0% loss -> Medium (6:1)
	bs, active = ctrl.UpdateLoss(5.0)
	if !active || bs != 6 {
		t.Errorf("expected Medium (6:1) at 5.0%% loss, got bs=%d active=%v", bs, active)
	}

	// 12.0% loss -> Aggressive (4:1)
	bs, active = ctrl.UpdateLoss(12.0)
	if !active || bs != 4 {
		t.Errorf("expected Aggressive (4:1) at 12.0%% loss, got bs=%d active=%v", bs, active)
	}

	// Disabled test
	ctrl.SetEnabled(false)
	bs, active = ctrl.UpdateLoss(12.0)
	if active || bs != 0 {
		t.Errorf("expected inactive when disabled, got bs=%d active=%v", bs, active)
	}
}

func TestEngineFECToggleAndStats(t *testing.T) {
	eng, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}

	if !eng.IsFECEnabled() {
		t.Errorf("expected FEC to be enabled by default")
	}

	stats := eng.Stats()
	if !stats.FECActive {
		t.Errorf("expected Stats.FECActive to be true")
	}
	if stats.FECRatio != "6:1" {
		t.Errorf("expected initial FECRatio 6:1, got %s", stats.FECRatio)
	}

	eng.SetFECEnabled(false)
	if eng.IsFECEnabled() {
		t.Errorf("expected FEC to be disabled")
	}

	stats = eng.Stats()
	if stats.FECActive {
		t.Errorf("expected Stats.FECActive to be false after toggle")
	}
	if stats.FECRatio != "Off" {
		t.Errorf("expected FECRatio 'Off' when disabled, got %s", stats.FECRatio)
	}
}
