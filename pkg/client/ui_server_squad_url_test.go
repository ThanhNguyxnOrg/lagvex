package client

import "testing"

// TestRelaySignalingURL covers the address forms a relay can report.
//
// A dual-stack dialer usually reports an IPv4-mapped IPv6 address; without
// unmapping it the derived URL is invalid and the squad feature silently
// degrades to a local-only room.
func TestRelaySignalingURL(t *testing.T) {
	cases := []struct {
		relayAddr string
		want      string
	}{
		{"127.0.0.1:4433", "http://127.0.0.1:4434"},
		{"[::ffff:127.0.0.1]:4433", "http://127.0.0.1:4434"},
		{"192.0.2.10:51820", "http://192.0.2.10:51821"},
		{"[2001:db8::1]:51820", "http://[2001:db8::1]:51821"},
		{"", ""},
		{"not-an-address", ""},
	}
	for _, tc := range cases {
		if got := relaySignalingURL(tc.relayAddr); got != tc.want {
			t.Errorf("relaySignalingURL(%q) = %q, want %q", tc.relayAddr, got, tc.want)
		}
	}
}
