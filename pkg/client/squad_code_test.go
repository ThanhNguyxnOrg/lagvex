package client

import (
	"strings"
	"testing"
)

// TestSquadRoomCodeIsUnguessable guards the room code generator: a room code
// grants access to the room roster, so it must not be derived from a clock
// value with a tiny code space.
func TestSquadRoomCodeIsUnguessable(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		code := newSquadRoomCode()
		if !strings.HasPrefix(code, "LGVX-") {
			t.Fatalf("room code %q lost the LGVX prefix", code)
		}
		if len(code) != len("LGVX-")+6 {
			t.Fatalf("room code %q has an unexpected length", code)
		}
		for _, r := range code[len("LGVX-"):] {
			if !strings.ContainsRune(squadRoomCodeAlphabet, r) {
				t.Fatalf("room code %q contains an unexpected character %q", code, r)
			}
		}
		if seen[code] {
			t.Fatalf("room code %q was generated twice", code)
		}
		seen[code] = true
	}
}
