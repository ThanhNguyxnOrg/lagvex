package protocol

import (
	"bytes"
	"testing"
)

func TestReedSolomonBurstLossRecovery(t *testing.T) {
	tests := []struct {
		name    string
		k       int
		m       int
		dropped []int // indices in [0, k) to drop simultaneously
	}{
		{name: "RS(6,2) recover 2 consecutive drops", k: 6, m: 2, dropped: []int{2, 3}},
		{name: "RS(6,2) recover first and last", k: 6, m: 2, dropped: []int{0, 5}},
		{name: "RS(4,2) recover 2 drops", k: 4, m: 2, dropped: []int{1, 2}},
		{name: "RS(4,3) recover 3 consecutive burst drops", k: 4, m: 3, dropped: []int{0, 1, 2}},
		{name: "RS(4,3) recover 3 scattered drops", k: 4, m: 3, dropped: []int{0, 2, 3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewFECEncoder(FECEncoderConfig{
				DefaultBlockSize:   tc.k,
				DefaultParityCount: tc.m,
			})
			dec := NewFECDecoder(128)

			orig := make([][]byte, tc.k)
			droppedSet := map[int]bool{}
			for _, d := range tc.dropped {
				droppedSet[d] = true
			}

			var emitted [][]byte
			baseSeq := uint64(100)
			for i := 0; i < tc.k; i++ {
				// Variable packet lengths (including short & long game ticks)
				pkt := make([]byte, 32+i*17)
				for b := range pkt {
					pkt[b] = byte((i*37 + b*13 + 7) & 0xff)
				}
				orig[i] = pkt

				pkts := enc.AddPacketMulti(baseSeq+uint64(i), pkt)
				emitted = append(emitted, pkts...)

				if !droppedSet[i] {
					dec.RecordPacket(baseSeq+uint64(i), pkt)
				}
			}

			if len(emitted) != tc.m {
				t.Fatalf("expected %d parity payloads, got %d", tc.m, len(emitted))
			}

			// Feed parity packets in reverse order to test out-of-order arrival.
			recovered := map[uint64][]byte{}
			for i := len(emitted) - 1; i >= 0; i-- {
				payload, err := DecodeFECPayload(emitted[i])
				if err != nil {
					t.Fatalf("DecodeFECPayload(%d): %v", i, err)
				}
				if int(payload.ParityCount) != tc.m {
					t.Fatalf("ParityCount = %d, want %d", payload.ParityCount, tc.m)
				}
				for _, rec := range dec.ProcessParityMulti(payload) {
					recovered[rec.Seq] = rec.Data
				}
			}

			if len(recovered) != len(tc.dropped) {
				t.Fatalf("recovered %d packets, want %d", len(recovered), len(tc.dropped))
			}
			for _, idx := range tc.dropped {
				seq := baseSeq + uint64(idx)
				got, ok := recovered[seq]
				if !ok {
					t.Fatalf("missing recovered packet for seq %d (index %d)", seq, idx)
				}
				if !bytes.Equal(got, orig[idx]) {
					t.Fatalf("recovered packet %d bytes mismatch (got len %d, want %d)", idx, len(got), len(orig[idx]))
				}
			}
		})
	}
}

func TestAdaptiveFECControllerRSProfiles(t *testing.T) {
	ctrl := NewAdaptiveFECController()

	k, m, _ := ctrl.UpdateLossProfile(0.1)
	if k != 0 || m != 0 {
		t.Fatalf("expected standby (0,0), got (%d,%d)", k, m)
	}
	k, m, _ = ctrl.UpdateLossProfile(1.5)
	if k != 10 || m != 1 {
		t.Fatalf("expected light XOR (10,1), got (%d,%d)", k, m)
	}
	k, m, _ = ctrl.UpdateLossProfile(5.0)
	if k != 6 || m != 2 {
		t.Fatalf("expected medium RS (6,2), got (%d,%d)", k, m)
	}
	k, m, _ = ctrl.UpdateLossProfile(9.5)
	if k != 4 || m != 2 {
		t.Fatalf("expected high RS (4,2), got (%d,%d)", k, m)
	}
	k, m, _ = ctrl.UpdateLossProfile(14.0)
	if k != 4 || m != 3 {
		t.Fatalf("expected extreme RS (4,3), got (%d,%d)", k, m)
	}
}
