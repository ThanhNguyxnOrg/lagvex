package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// TCP ping limits. The tool exists to measure the handshake RTT to a game
// server the user names, not to scan the internet: it accepts IP literals
// only, refuses non-public addresses, sends at most eight SYNs per run and is
// rate-limited per process.
const (
	tcpPingDefaultPort     = 443
	tcpPingDefaultAttempts = 4
	tcpPingMaxAttempts     = 8
	tcpPingDefaultTimeout  = 1200 * time.Millisecond
	tcpPingMinTimeout      = 200 * time.Millisecond
	tcpPingMaxTimeout      = 3000 * time.Millisecond
	tcpPingDefaultInterval = 250 * time.Millisecond
	tcpPingMinInterval     = 100 * time.Millisecond
	tcpPingMaxInterval     = time.Second
	tcpPingCooldown        = 2 * time.Second
	tcpPingHourlyLimit     = 60
)

// Non-public IPv4 ranges that net.IP.IsPrivate does not cover but which must
// never be dialed by a user-facing measurement tool: "this network", carrier
// NAT, IETF protocol assignments, benchmarking and reserved space.
var tcpPingBlockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
}

// TCPPingSample is one TCP handshake attempt.
type TCPPingSample struct {
	Attempt int     `json:"attempt"`
	RTTMs   float64 `json:"rttMs"`
	// Outcome is "open" (handshake completed), "closed" (host answered with
	// TCP RST), "timeout" (no answer before the deadline) or "error".
	Outcome string `json:"outcome"`
	Error   string `json:"error,omitempty"`
}

// TCPPingResult is a complete, measured TCP ping run. A closed port still
// yields a valid RTT because the RST proves the host is reachable.
type TCPPingResult struct {
	Target     string          `json:"target"`
	Port       int             `json:"port"`
	Attempts   int             `json:"attempts"`
	TimeoutMs  int             `json:"timeoutMs"`
	Measured   bool            `json:"measured"`
	MinMs      float64         `json:"minMs"`
	AvgMs      float64         `json:"avgMs"`
	MaxMs      float64         `json:"maxMs"`
	LossPct    float64         `json:"lossPct"`
	Samples    []TCPPingSample `json:"samples"`
	StartedAt  string          `json:"startedAt"`
	FinishedAt string          `json:"finishedAt"`
	Note       string          `json:"note"`
}

// validateTCPPingTarget accepts only an explicit public unicast IP address.
// Hostnames are rejected so the endpoint cannot be used as a DNS-driven scan
// proxy, and private/loopback/link-local ranges are rejected outright.
func validateTCPPingTarget(raw string) (net.IP, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, errors.New("target IP is required")
	}
	addr, err := netip.ParseAddr(trimmed)
	if err != nil {
		return nil, errors.New("target must be an IP address (hostnames are not resolved)")
	}
	addr = addr.Unmap()
	if !isPublicTCPPingAddr(addr) {
		return nil, errors.New("target must be a public unicast IP address")
	}
	return net.IP(addr.AsSlice()), nil
}

func isPublicTCPPingAddr(addr netip.Addr) bool {
	if !addr.IsValid() || !addr.IsGlobalUnicast() {
		return false
	}
	if addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() ||
		addr.IsInterfaceLocalMulticast() {
		return false
	}
	if addr.Is4() {
		for _, prefix := range tcpPingBlockedPrefixes {
			if prefix.Contains(addr) {
				return false
			}
		}
	}
	return true
}

// runTCPPing performs the measured handshakes. The caller validates the
// address first; tests call it directly against a loopback listener.
func runTCPPing(ctx context.Context, ip net.IP, port, attempts int, timeout, interval time.Duration) TCPPingResult {
	started := time.Now()
	result := TCPPingResult{
		Target:    ip.String(),
		Port:      port,
		Attempts:  attempts,
		TimeoutMs: int(timeout / time.Millisecond),
		Samples:   make([]TCPPingSample, 0, attempts),
		Note:      "TCP handshake RTT (SYN to SYN-ACK); no application data is sent.",
	}

	var rtts []float64
	closed := 0
	failed := 0
	dialer := net.Dialer{Timeout: timeout}

	for i := 1; i <= attempts; i++ {
		if i > 1 {
			select {
			case <-ctx.Done():
				break
			case <-time.After(interval):
			}
		}
		if ctx.Err() != nil {
			break
		}

		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
		rtt := roundTo(float64(time.Since(start).Microseconds())/1000.0, 2)

		sample := TCPPingSample{Attempt: i}
		switch {
		case err == nil:
			_ = conn.Close()
			sample.RTTMs = rtt
			sample.Outcome = "open"
			rtts = append(rtts, rtt)
		case errors.Is(err, syscall.ECONNREFUSED):
			// A RST is a response: the path RTT is real even though the port
			// is closed. Count it as measured and say so in the note.
			sample.RTTMs = rtt
			sample.Outcome = "closed"
			rtts = append(rtts, rtt)
			closed++
		default:
			sample.Outcome = tcpPingOutcome(err)
			sample.Error = err.Error()
			failed++
		}
		result.Samples = append(result.Samples, sample)
	}

	result.StartedAt = started.UTC().Format(time.RFC3339Nano)
	result.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)

	if len(rtts) > 0 {
		result.Measured = true
		result.MinMs = rtts[0]
		result.MaxMs = rtts[0]
		sum := 0.0
		for _, rtt := range rtts {
			sum += rtt
			if rtt < result.MinMs {
				result.MinMs = rtt
			}
			if rtt > result.MaxMs {
				result.MaxMs = rtt
			}
		}
		result.AvgMs = roundTo(sum/float64(len(rtts)), 2)
	}
	if attempts > 0 {
		result.LossPct = roundTo(float64(failed)/float64(attempts)*100, 1)
	}

	var notes []string
	notes = append(notes, result.Note)
	if closed > 0 {
		notes = append(notes, fmt.Sprintf("Port %d is closed; the host answered with TCP RST, so those handshake times are still real path measurements.", port))
	}
	if failed > 0 {
		notes = append(notes, fmt.Sprintf("%d of %d attempts produced no TCP response; loss counts only attempts without any response.", failed, attempts))
	}
	result.Note = strings.Join(notes, " ")
	return result
}

func tcpPingOutcome(err error) string {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	return "error"
}

// tcpPingLimiter bounds how often the dashboard can run a TCP ping. now is
// injectable so tests can exercise the hourly window without sleeping.
type tcpPingLimiter struct {
	mu          sync.Mutex
	now         func() time.Time
	lastRun     time.Time
	windowStart time.Time
	windowCount int
}

// allow returns a retry delay and an error when the run must be refused.
func (l *tcpPingLimiter) allow() (time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if l.now != nil {
		now = l.now()
	}
	if l.windowStart.IsZero() || now.Sub(l.windowStart) >= time.Hour {
		l.windowStart = now
		l.windowCount = 0
	}
	if !l.lastRun.IsZero() {
		if wait := tcpPingCooldown - now.Sub(l.lastRun); wait > 0 {
			return wait, fmt.Errorf("wait %s before the next TCP ping", wait.Round(time.Millisecond))
		}
	}
	if l.windowCount >= tcpPingHourlyLimit {
		retry := l.windowStart.Add(time.Hour).Sub(now)
		if retry < 0 {
			retry = 0
		}
		return retry, fmt.Errorf("hourly TCP ping limit reached (%d runs)", tcpPingHourlyLimit)
	}

	l.lastRun = now
	l.windowCount++
	return 0, nil
}

// clampTCPPingDefaults fills omitted request fields and validates ranges.
func clampTCPPingDefaults(port, attempts, timeoutMs int) (int, int, time.Duration, error) {
	if port == 0 {
		port = tcpPingDefaultPort
	}
	if port < 1 || port > 65535 {
		return 0, 0, 0, errors.New("port must be between 1 and 65535")
	}
	if attempts == 0 {
		attempts = tcpPingDefaultAttempts
	}
	if attempts < 1 || attempts > tcpPingMaxAttempts {
		return 0, 0, 0, fmt.Errorf("attempts must be between 1 and %d", tcpPingMaxAttempts)
	}
	timeout := tcpPingDefaultTimeout
	if timeoutMs != 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	if timeout < tcpPingMinTimeout || timeout > tcpPingMaxTimeout {
		return 0, 0, 0, fmt.Errorf("timeoutMs must be between %d and %d", tcpPingMinTimeout/time.Millisecond, tcpPingMaxTimeout/time.Millisecond)
	}
	return port, attempts, timeout, nil
}

// retryAfterSeconds converts a limiter delay into a whole-second header value.
func retryAfterSeconds(wait time.Duration) int {
	seconds := int(math.Ceil(wait.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return seconds
}
