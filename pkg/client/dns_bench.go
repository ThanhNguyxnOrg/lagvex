package client

import (
	"context"
	"encoding/binary"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

// DNSResolverSpec defines a public DNS resolver to benchmark.
type DNSResolverSpec struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Category string `json:"category"`
}

// DNSBenchmarkEntry is the measured latency result for one DNS resolver.
type DNSBenchmarkEntry struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Category  string  `json:"category"`
	MedianMs  float64 `json:"medianMs,omitempty"`
	MinMs     float64 `json:"minMs,omitempty"`
	MaxMs     float64 `json:"maxMs,omitempty"`
	LossPct   float64 `json:"lossPct"`
	Samples   int     `json:"samples"`
	Reachable bool    `json:"reachable"`
	IsFastest bool    `json:"isFastest"`
}

var defaultDNSResolvers = []DNSResolverSpec{
	{ID: "cloudflare", Name: "Cloudflare", Address: "1.1.1.1:53", Category: "Low-Latency"},
	{ID: "cloudflare-sec", Name: "Cloudflare Security", Address: "1.1.1.2:53", Category: "Malware Filter"},
	{ID: "google", Name: "Google Public DNS", Address: "8.8.8.8:53", Category: "Global Anycast"},
	{ID: "quad9", Name: "Quad9", Address: "9.9.9.9:53", Category: "Security"},
	{ID: "opendns", Name: "Cisco OpenDNS", Address: "208.67.222.222:53", Category: "Global Anycast"},
	{ID: "adguard", Name: "AdGuard DNS", Address: "94.140.14.14:53", Category: "Ad & Tracker Block"},
}

// buildDNSQuery constructs a minimal RFC 1035 standard A-record query packet.
func buildDNSQuery(txID uint16, domain string) []byte {
	var buf []byte
	hdr := make([]byte, 12)
	binary.BigEndian.PutUint16(hdr[0:2], txID)
	binary.BigEndian.PutUint16(hdr[2:4], 0x0100) // RD = 1 (recursion desired)
	binary.BigEndian.PutUint16(hdr[4:6], 1)      // QDCOUNT = 1
	buf = append(buf, hdr...)

	for _, label := range strings.Split(strings.TrimSuffix(domain, "."), ".") {
		if len(label) == 0 || len(label) > 63 {
			continue
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, label...)
	}
	buf = append(buf, 0x00)       // root label
	buf = append(buf, 0x00, 0x01) // QTYPE = A
	buf = append(buf, 0x00, 0x01) // QCLASS = IN
	return buf
}

func benchmarkSingleResolver(ctx context.Context, spec DNSResolverSpec, samples int) DNSBenchmarkEntry {
	if samples <= 0 {
		samples = 3
	}
	entry := DNSBenchmarkEntry{
		ID:       spec.ID,
		Name:     spec.Name,
		Address:  strings.TrimSuffix(spec.Address, ":53"),
		Category: spec.Category,
		Samples:  samples,
	}

	var rtts []float64
	respBuf := make([]byte, 512)
	for i := 0; i < samples; i++ {
		if ctx.Err() != nil {
			break
		}
		txID := uint16(0x4C00 + i)
		query := buildDNSQuery(txID, "cloudflare.com")

		dialer := net.Dialer{Timeout: 800 * time.Millisecond}
		conn, err := dialer.DialContext(ctx, "udp4", spec.Address)
		if err != nil {
			continue
		}
		deadline := time.Now().Add(900 * time.Millisecond)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		_ = conn.SetDeadline(deadline)

		start := time.Now()
		_, wErr := conn.Write(query)
		if wErr == nil {
			n, rErr := conn.Read(respBuf)
			if rErr == nil && n >= 12 && binary.BigEndian.Uint16(respBuf[0:2]) == txID {
				rtts = append(rtts, float64(time.Since(start).Microseconds())/1000.0)
			}
		}
		_ = conn.Close()
		time.Sleep(25 * time.Millisecond)
	}

	lost := samples - len(rtts)
	if lost < 0 {
		lost = 0
	}
	entry.LossPct = roundTo(float64(lost)/float64(samples)*100.0, 1)

	if len(rtts) > 0 {
		sort.Float64s(rtts)
		entry.Reachable = true
		entry.MinMs = roundTo(rtts[0], 2)
		entry.MaxMs = roundTo(rtts[len(rtts)-1], 2)
		entry.MedianMs = roundTo(calculateMedian(rtts), 2)
	}
	return entry
}

// RunDNSBenchmark measures real UDP lookup latency across resolvers concurrently.
func RunDNSBenchmark(ctx context.Context, resolvers []DNSResolverSpec, samples int) []DNSBenchmarkEntry {
	if len(resolvers) == 0 {
		resolvers = defaultDNSResolvers
	}
	results := make([]DNSBenchmarkEntry, len(resolvers))
	var wg sync.WaitGroup
	for i, spec := range resolvers {
		wg.Add(1)
		go func(idx int, s DNSResolverSpec) {
			defer wg.Done()
			results[idx] = benchmarkSingleResolver(ctx, s, samples)
		}(i, spec)
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Reachable != results[j].Reachable {
			return results[i].Reachable
		}
		if results[i].LossPct != results[j].LossPct {
			return results[i].LossPct < results[j].LossPct
		}
		return results[i].MedianMs < results[j].MedianMs
	})
	for i := range results {
		if results[i].Reachable {
			results[i].IsFastest = true
			break
		}
	}
	return results
}
