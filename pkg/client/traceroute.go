package client

import (
	"context"
	"fmt"
	"net/netip"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// TraceHop is one measured network hop along the path to a public IPv4 target.
type TraceHop struct {
	Hop     int       `json:"hop"`
	IP      string    `json:"ip,omitempty"`
	RTTsMs  []float64 `json:"rttsMs,omitempty"`
	AvgMs   float64   `json:"avgMs,omitempty"`
	MinMs   float64   `json:"minMs,omitempty"`
	MaxMs   float64   `json:"maxMs,omitempty"`
	LossPct float64   `json:"lossPct"`
	Timeout bool      `json:"timeout"`
}

// TracerouteResult holds the structured output of a PingPlotter-style hop scan.
type TracerouteResult struct {
	Host       string     `json:"host"`
	MaxHops    int        `json:"maxHops"`
	Tool       string     `json:"tool"`
	Hops       []TraceHop `json:"hops"`
	MeasuredAt string     `json:"measuredAt"`
	DurationMs int64      `json:"durationMs"`
	Error      string     `json:"error,omitempty"`
}

var (
	reIPv4Token = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	reRTTToken  = regexp.MustCompile(`(?:<)?(\d+(?:\.\d+)?)\s*ms`)
)

// runTraceroute executes the platform's native numeric traceroute binary
// against a validated public IPv4 literal.
func runTraceroute(ctx context.Context, ip netip.Addr, maxHops int) TracerouteResult {
	if maxHops < 4 {
		maxHops = 4
	}
	if maxHops > 30 {
		maxHops = 30
	}

	start := time.Now()
	res := TracerouteResult{
		Host:       ip.String(),
		MaxHops:    maxHops,
		Hops:       []TraceHop{},
		MeasuredAt: time.Now().UTC().Format(time.RFC3339),
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		res.Tool = "tracert"
		cmd = execHiddenContext(ctx, "tracert", "-d", "-w", "1000", "-h", strconv.Itoa(maxHops), ip.String())
	} else if path, err := exec.LookPath("traceroute"); err == nil {
		res.Tool = "traceroute"
		cmd = execHiddenContext(ctx, path, "-n", "-q", "2", "-w", "1", "-m", strconv.Itoa(maxHops), ip.String())
	} else if path, err := exec.LookPath("tracepath"); err == nil {
		res.Tool = "tracepath"
		cmd = execHiddenContext(ctx, path, "-n", "-m", strconv.Itoa(maxHops), ip.String())
	} else {
		res.Error = "no traceroute or tracepath binary found on this system"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	out, err := cmd.CombinedOutput()
	res.DurationMs = time.Since(start).Milliseconds()

	hops := parseTracerouteOutput(string(out))
	res.Hops = hops
	if len(hops) == 0 && err != nil {
		res.Error = fmt.Sprintf("traceroute failed: %v", err)
	}
	return res
}

// parseTracerouteOutput parses numeric output from Unix traceroute, Linux
// tracepath, and Windows tracert into structured hop telemetry.
func parseTracerouteOutput(raw string) []TraceHop {
	var hops []TraceHop
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Hop number may have a trailing colon in tracepath ("1:" or "1?:").
		hopStr := strings.TrimRight(fields[0], ":?")
		hopNum, err := strconv.Atoi(hopStr)
		if err != nil || hopNum <= 0 || hopNum > 64 {
			continue
		}

		body := strings.TrimSpace(line[len(fields[0]):])
		ipMatch := reIPv4Token.FindString(body)

		var rtts []float64
		for _, m := range reRTTToken.FindAllStringSubmatch(body, -1) {
			if len(m) >= 2 {
				if v, err := strconv.ParseFloat(m[1], 64); err == nil {
					rtts = append(rtts, roundTo(v, 2))
				}
			}
		}

		stars := 0
		for _, tok := range strings.Fields(body) {
			if tok == "*" {
				stars++
			}
		}

		hop := TraceHop{
			Hop:    hopNum,
			IP:     ipMatch,
			RTTsMs: rtts,
		}

		if len(rtts) == 0 {
			hop.Timeout = true
			hop.LossPct = 100.0
		} else {
			minV, maxV, sum := rtts[0], rtts[0], 0.0
			for _, v := range rtts {
				if v < minV {
					minV = v
				}
				if v > maxV {
					maxV = v
				}
				sum += v
			}
			hop.MinMs = roundTo(minV, 2)
			hop.MaxMs = roundTo(maxV, 2)
			hop.AvgMs = roundTo(sum/float64(len(rtts)), 2)
			totalProbes := len(rtts) + stars
			if totalProbes > 0 {
				hop.LossPct = roundTo(float64(stars)/float64(totalProbes)*100.0, 1)
			}
		}
		hops = append(hops, hop)
	}
	return hops
}
