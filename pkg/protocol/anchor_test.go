package protocol

import "testing"

func TestAnchorProbeRequestRoundTrip(t *testing.T) {
	packed := EncodeAnchorProbeRequest(0xDEADBEEF, 7)
	seq, anchorID := DecodeAnchorProbeRequest(packed)
	if seq != 0xDEADBEEF {
		t.Fatalf("seq = %#x, want 0xDEADBEEF", seq)
	}
	if anchorID != 7 {
		t.Fatalf("anchorID = %d, want 7", anchorID)
	}
}

func TestAnchorProbeReplyRoundTrip(t *testing.T) {
	packed := EncodeAnchorProbeReply(12345, 42_500, 50)
	seq, rttUs, lossPct := DecodeAnchorProbeReply(packed)
	if seq != 12345 {
		t.Fatalf("seq = %d, want 12345", seq)
	}
	if rttUs != 42_500 {
		t.Fatalf("rttUs = %d, want 42500", rttUs)
	}
	if lossPct != 50 {
		t.Fatalf("lossPct = %d, want 50", lossPct)
	}
}

func TestAnchorProbeReplyClamps(t *testing.T) {
	packed := EncodeAnchorProbeReply(1, 0xFFFFFFFF, 250)
	_, rttUs, lossPct := DecodeAnchorProbeReply(packed)
	if rttUs != 0xFFFFFF {
		t.Fatalf("rttUs = %d, want the 24-bit cap", rttUs)
	}
	if lossPct != 100 {
		t.Fatalf("lossPct = %d, want 100", lossPct)
	}
}

func TestAnchorProbeTargetsAllowlist(t *testing.T) {
	if len(AnchorProbeTargets) == 0 {
		t.Fatalf("expected a non-empty allowlist")
	}
	for id, target := range AnchorProbeTargets {
		if id == 0 {
			t.Fatalf("anchor id 0 is reserved")
		}
		if target == "" {
			t.Fatalf("anchor %d has an empty target", id)
		}
	}
	if _, ok := AnchorProbeTargets[200]; ok {
		t.Fatalf("unexpected allowlist entry for id 200")
	}
}
