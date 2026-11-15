package client

import (
	"strings"
	"testing"
)

func TestDecodeProcNetAddressIPv4(t *testing.T) {
	got, ok := decodeProcNetAddress("5DB8D822:1F90", false)
	if !ok {
		t.Fatalf("expected a valid IPv4 address")
	}
	if got != "34.216.184.93:8080" {
		t.Fatalf("decoded %q, want 34.216.184.93:8080", got)
	}

	if _, ok := decodeProcNetAddress("not-hex:1F90", false); ok {
		t.Fatalf("malformed address must not decode")
	}
}

func TestDecodeProcNetAddressIPv6(t *testing.T) {
	got, ok := decodeProcNetAddress("B80D0120000000000000000001000000:0BB8", true)
	if !ok {
		t.Fatalf("expected a valid IPv6 address")
	}
	if got != "[2001:db8::1]:3000" {
		t.Fatalf("decoded %q, want [2001:db8::1]:3000", got)
	}
}

func TestParseProcNetTableSkipsListeners(t *testing.T) {
	table := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:BF68 5DB8D822:1F90 01 00000000:00000000 02:00000000 00000000  1000        0 424242 2 0000000000000000 20 4 30 10 -1
   1: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 424243 2 0000000000000000 100 0 0 10 0
`
	rows := parseProcNetTable(strings.NewReader(table), "tcp", false)
	if len(rows) != 1 {
		t.Fatalf("expected exactly one connected row, got %d: %+v", len(rows), rows)
	}

	conn, ok := rows["424242"]
	if !ok {
		t.Fatalf("connected row was not keyed by its inode: %+v", rows)
	}
	if conn.Remote != "34.216.184.93:8080" {
		t.Errorf("remote = %q, want 34.216.184.93:8080", conn.Remote)
	}
	if conn.Local != "127.0.0.1:49000" {
		t.Errorf("local = %q, want 127.0.0.1:49000", conn.Local)
	}
	if conn.State != "established" {
		t.Errorf("state = %q, want established", conn.State)
	}
	if conn.Protocol != "tcp" || conn.Family != "ipv4" {
		t.Errorf("unexpected protocol/family: %s/%s", conn.Protocol, conn.Family)
	}
}

func TestParseLsofConnections(t *testing.T) {
	output := `COMMAND   PID USER   FD   TYPE  DEVICE SIZE/OFF NODE NAME
game.exe  123 user   10u  IPv4 0x1234      0t0  UDP 192.168.1.7:5353->203.0.113.7:8080
game.exe  123 user   11u  IPv4 0x1234      0t0  TCP 192.168.1.7:55000->1.2.3.4:443 (ESTABLISHED)
game.exe  123 user   12u  IPv4 0x1234      0t0  TCP *:1900 (LISTEN)
`
	conns := parseLsofConnections(output)
	if len(conns) != 2 {
		t.Fatalf("expected 2 connected sockets, got %d: %+v", len(conns), conns)
	}

	var udp, tcp *ProcessConnection
	for i := range conns {
		switch conns[i].Protocol {
		case "udp":
			udp = &conns[i]
		case "tcp":
			tcp = &conns[i]
		}
	}
	if udp == nil || udp.Remote != "203.0.113.7:8080" {
		t.Fatalf("udp connection parsed incorrectly: %+v", conns)
	}
	if tcp == nil || tcp.Remote != "1.2.3.4:443" || tcp.State != "established" {
		t.Fatalf("tcp connection parsed incorrectly: %+v", conns)
	}
}

func TestMergeConnectionsDedupes(t *testing.T) {
	first := []ProcessConnection{
		{Protocol: "udp", Local: "127.0.0.1:1000", Remote: "1.2.3.4:5"},
	}
	second := []ProcessConnection{
		{Protocol: "udp", Local: "127.0.0.1:2000", Remote: "1.2.3.4:5"},
		{Protocol: "udp", Local: "127.0.0.1:3000"},
		{Protocol: "udp", Local: "127.0.0.1:3000"},
	}

	merged := mergeConnections(first, second)
	if len(merged) != 2 {
		t.Fatalf("expected 2 unique connections, got %d: %+v", len(merged), merged)
	}
	var hasPeer, hasLocalOnly bool
	for _, conn := range merged {
		if conn.Remote == "1.2.3.4:5" {
			hasPeer = true
		}
		if conn.Remote == "" && conn.Local == "127.0.0.1:3000" {
			hasLocalOnly = true
		}
	}
	if !hasPeer || !hasLocalOnly {
		t.Fatalf("unexpected merge result: %+v", merged)
	}
}

func TestFilterPublicPeerIPs(t *testing.T) {
	conns := []ProcessConnection{
		{Protocol: "udp", Remote: "203.0.113.7:8080"},
		{Protocol: "udp", Remote: "203.0.113.7:9090"}, // duplicate peer, different port
		{Protocol: "tcp", Remote: "198.51.100.4:443"},
		{Protocol: "udp", Remote: "192.168.1.10:5000"}, // private
		{Protocol: "udp", Remote: "127.0.0.1:5353"},    // loopback
		{Protocol: "udp", Remote: "100.64.1.2:1234"},   // CGNAT
		{Protocol: "udp", Remote: "169.254.1.1:1"},     // link-local
		{Protocol: "tcp", Remote: "[2001:db8::1]:443"}, // IPv6
		{Protocol: "udp", Local: "127.0.0.1:3000"},     // local-only, no peer
	}

	got := filterPublicPeerIPs(conns)
	want := []string{"198.51.100.4", "203.0.113.7"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
