package protocol

// AnchorProbeTargets is the fixed allowlist of public endpoints the relay may
// probe on behalf of a client. A client references an id; the relay never
// dials an arbitrary host, which keeps a Lagvex relay from being usable as a
// port scanner or DDoS reflector. Keep these ids in sync with the client's
// telemetry anchors.
var AnchorProbeTargets = map[byte]string{
	1:  "ec2.ap-southeast-1.amazonaws.com:443", // Singapore
	2:  "ec2.ap-northeast-1.amazonaws.com:443", // Tokyo
	3:  "ec2.ap-northeast-2.amazonaws.com:443", // Seoul
	4:  "ec2.ap-east-1.amazonaws.com:443",      // Hong Kong
	5:  "ec2.ap-south-1.amazonaws.com:443",     // Mumbai
	6:  "ec2.ap-southeast-2.amazonaws.com:443", // Sydney
	7:  "ec2.us-east-1.amazonaws.com:443",      // N. Virginia
	8:  "ec2.us-west-1.amazonaws.com:443",      // N. California
	9:  "ec2.eu-central-1.amazonaws.com:443",   // Frankfurt
	10: "ec2.sa-east-1.amazonaws.com:443",      // São Paulo
	11: "ec2.af-south-1.amazonaws.com:443",     // Cape Town
	12: "ec2.eu-west-3.amazonaws.com:443",      // Paris
	13: "ec2.eu-west-2.amazonaws.com:443",      // London
	14: "ec2.me-south-1.amazonaws.com:443",     // Bahrain
	15: "ec2.ap-northeast-3.amazonaws.com:443", // Osaka
}

// EncodeAnchorProbeRequest packs a sequence number and an allowlisted anchor
// id into the 8-byte control payload.
func EncodeAnchorProbeRequest(seq uint32, anchorID byte) uint64 {
	return uint64(seq)<<8 | uint64(anchorID)
}

// DecodeAnchorProbeRequest unpacks a sequence number and anchor id.
func DecodeAnchorProbeRequest(payload uint64) (seq uint32, anchorID byte) {
	return uint32(payload >> 8), byte(payload)
}

// EncodeAnchorProbeReply packs the matching sequence number, the relay-measured
// RTT in microseconds (24 bits, capped) and loss percent (0-100).
func EncodeAnchorProbeReply(seq uint32, rttUs uint32, lossPct uint8) uint64 {
	if rttUs > 0xFFFFFF {
		rttUs = 0xFFFFFF
	}
	if lossPct > 100 {
		lossPct = 100
	}
	return uint64(seq)<<32 | uint64(rttUs)<<8 | uint64(lossPct)
}

// DecodeAnchorProbeReply unpacks a relay probe result.
func DecodeAnchorProbeReply(payload uint64) (seq uint32, rttUs uint32, lossPct uint8) {
	return uint32(payload >> 32), uint32(payload>>8) & 0xFFFFFF, uint8(payload)
}
