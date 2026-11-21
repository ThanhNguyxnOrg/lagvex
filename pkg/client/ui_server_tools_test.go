package client

import (
	"context"
	"encoding/binary"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestParseTracerouteOutputUnixAndWindows(t *testing.T) {
	unixOut := `
traceroute to 1.1.1.1 (1.1.1.1), 18 hops max, 60 byte packets
 1  192.168.1.1  1.123 ms  1.456 ms
 2  * *
 3  103.1.2.3  8.500 ms  *
 4  1.1.1.1  14.200 ms  15.800 ms
`
	hops := parseTracerouteOutput(unixOut)
	if len(hops) != 4 {
		t.Fatalf("expected 4 parsed hops, got %d: %+v", len(hops), hops)
	}
	if hops[0].IP != "192.168.1.1" || hops[0].LossPct != 0 || hops[0].Timeout {
		t.Fatalf("hop 1 unexpected: %+v", hops[0])
	}
	if !hops[1].Timeout || hops[1].LossPct != 100 {
		t.Fatalf("hop 2 expected timeout/100%% loss: %+v", hops[1])
	}
	if hops[2].LossPct != 50 || hops[2].AvgMs != 8.5 {
		t.Fatalf("hop 3 expected 50%% loss / 8.5ms avg: %+v", hops[2])
	}
	if hops[3].IP != "1.1.1.1" || hops[3].AvgMs != 15 {
		t.Fatalf("hop 4 unexpected: %+v", hops[3])
	}

	winOut := `
Tracing route to 1.1.1.1 over a maximum of 18 hops

  1    <1 ms     1 ms     1 ms  192.168.1.1
  2     *        *        *     Request timed out.
  3    12 ms    14 ms    13 ms  1.1.1.1
`
	winHops := parseTracerouteOutput(winOut)
	if len(winHops) != 3 {
		t.Fatalf("expected 3 windows hops, got %d: %+v", len(winHops), winHops)
	}
	if winHops[0].IP != "192.168.1.1" || winHops[0].AvgMs != 1 {
		t.Fatalf("win hop 1 unexpected: %+v", winHops[0])
	}
	if !winHops[1].Timeout {
		t.Fatalf("win hop 2 expected timeout: %+v", winHops[1])
	}
	if winHops[2].IP != "1.1.1.1" || winHops[2].AvgMs != 13 {
		t.Fatalf("win hop 3 unexpected: %+v", winHops[2])
	}
}

func TestTracerouteRejectsPrivateIPs(t *testing.T) {
	u := newTestShareServer(t)
	h := withLocalRequestGuard(u.setupRoutes())

	for _, bad := range []string{"127.0.0.1", "192.168.1.1", "10.0.0.1", "google.com"} {
		rec := doGuardedPost(t, h, "/api/traceroute", map[string]any{"host": bad})
		if rec.Code != http.StatusBadRequest {
			t.Errorf("host %q: status = %d, want 400 (body: %s)", bad, rec.Code, rec.Body.String())
		}
	}
}

func TestRunDNSBenchmarkAgainstLocalUDPServer(t *testing.T) {
	// Spin up a minimal mock UDP DNS responder on loopback.
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer func() { _ = pc.Close() }()

	go func() {
		buf := make([]byte, 512)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			if n < 12 {
				continue
			}
			resp := make([]byte, 12)
			binary.BigEndian.PutUint16(resp[0:2], binary.BigEndian.Uint16(buf[0:2]))
			binary.BigEndian.PutUint16(resp[2:4], 0x8180)
			_, _ = pc.WriteTo(resp, addr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res := RunDNSBenchmark(ctx, []DNSResolverSpec{
		{ID: "mock", Name: "Mock DNS", Address: pc.LocalAddr().String(), Category: "Test"},
	}, 2)
	if len(res) != 1 || !res[0].Reachable || !res[0].IsFastest || res[0].LossPct != 0 {
		t.Fatalf("unexpected DNS benchmark result: %+v", res)
	}
}
