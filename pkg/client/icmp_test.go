package client

import (
	"testing"
)

func TestICMPChecksumKnownVector(t *testing.T) {
	// RFC 1071 style check: checksum of a fixed message must be stable and
	// must validate to zero when the checksum field is included.
	msg := buildICMPEchoRequest(0x1234, 7)
	if msg[0] != 8 || msg[1] != 0 {
		t.Fatalf("unexpected ICMP type/code: %d/%d", msg[0], msg[1])
	}
	if icmpChecksum(msg) != 0 {
		t.Fatalf("a message carrying its own checksum must validate to zero")
	}
}

func TestMatchICMPEchoReply(t *testing.T) {
	request := buildICMPEchoRequest(1, 42)

	// A valid reply for seq 42 matches.
	reply := make([]byte, len(request))
	copy(reply, request)
	reply[0] = 0 // echo reply
	reply[2], reply[3] = 0, 0
	cs := icmpChecksum(reply)
	reply[2], reply[3] = byte(cs>>8), byte(cs)
	if !matchICMPEchoReply(reply, 42) {
		t.Fatalf("expected the echo reply for seq 42 to match")
	}

	// Wrong sequence, wrong type and truncated messages must not match.
	if matchICMPEchoReply(reply, 43) {
		t.Fatalf("a reply with a different sequence must not match")
	}
	reply[0] = 8
	if matchICMPEchoReply(reply, 42) {
		t.Fatalf("an echo request must not be treated as a reply")
	}
	if matchICMPEchoReply(reply[:10], 42) {
		t.Fatalf("a truncated message must not match")
	}
}
