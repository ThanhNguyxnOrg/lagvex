package client

// icmpProbeResult is one ICMP echo measurement to a relay host.
type icmpProbeResult struct {
	RTTMs   float64
	LossPct float64
	Samples int
}

// icmpChecksum computes the RFC 1071 internet checksum used by ICMP echo
// messages. It is exported to tests within the package.
func icmpChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

// buildICMPEchoRequest builds an ICMPv4 echo request with an 8-byte payload
// carrying a sentinel so replies can be matched safely.
func buildICMPEchoRequest(id, seq uint16) []byte {
	msg := make([]byte, 16)
	msg[0] = 8 // echo request
	msg[1] = 0 // code
	msg[4] = byte(id >> 8)
	msg[5] = byte(id)
	msg[6] = byte(seq >> 8)
	msg[7] = byte(seq)
	copy(msg[8:], []byte{'l', 'a', 'g', 'v', 'e', 'x', 0x01, 0x02})
	cs := icmpChecksum(msg)
	msg[2] = byte(cs >> 8)
	msg[3] = byte(cs)
	return msg
}

// matchICMPEchoReply reports whether an ICMP message is the echo reply for the
// request with the given sequence number and sentinel payload.
func matchICMPEchoReply(msg []byte, seq uint16) bool {
	if len(msg) < 16 {
		return false
	}
	if msg[0] != 0 || msg[1] != 0 { // echo reply, code 0
		return false
	}
	if uint16(msg[6])<<8|uint16(msg[7]) != seq {
		return false
	}
	return string(msg[8:16]) == "lagvex\x01\x02"
}
