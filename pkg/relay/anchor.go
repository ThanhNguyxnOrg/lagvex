package relay

import (
	"context"
	"net"
	"sort"
	"time"
)

// anchorProbeSamples is deliberately small: two TCP handshakes give a median
// without making the relay a useful amplifier.
const anchorProbeSamples = 2

// probeAnchorTarget performs TCP handshakes to one allowlisted anchor and
// returns the median RTT in microseconds plus the loss percent.
func probeAnchorTarget(target string) (uint32, uint8) {
	rtts := make([]time.Duration, 0, anchorProbeSamples)
	for i := 0; i < anchorProbeSamples; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		dialer := net.Dialer{}
		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", target)
		cancel()
		if err != nil {
			continue
		}
		rtts = append(rtts, time.Since(start))
		_ = conn.Close()
	}

	if len(rtts) == 0 {
		return 0, 100
	}
	sort.Slice(rtts, func(i, j int) bool { return rtts[i] < rtts[j] })
	median := rtts[len(rtts)/2]
	lossPct := uint8((anchorProbeSamples - len(rtts)) * 100 / anchorProbeSamples)
	return uint32(median.Microseconds()), lossPct
}
