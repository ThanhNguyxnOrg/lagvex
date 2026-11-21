package relay

import (
	"net/netip"
	"testing"
)

func TestValidatePSK(t *testing.T) {
	for _, tc := range []struct {
		name    string
		psk     string
		wantErr bool
	}{
		{"empty", "", true},
		{"old Compose default", "change_this_to_your_secure_psk", true},
		{"operator supplied", "test-only-key", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				ListenAddr: ":51820",
				Subnet:     netip.MustParsePrefix("10.88.0.0/24"),
				PSK:        []byte(tc.psk),
			}
			if err := cfg.Validate(); (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}
