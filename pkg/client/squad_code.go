package client

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// squadRoomCodeAlphabet omits characters that are easy to confuse when a room
// code is read aloud or typed from a screenshot.
const squadRoomCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// newSquadRoomCode returns a cryptographically random room code such as
// "LGVX-K7P2QD".
//
// Room codes are secrets: they grant access to the room roster, so they must
// not be guessable. The old scheme derived the code from a clock value with
// only 9000 possible values, which could be enumerated. When the relay is
// reachable it generates the code itself; this helper covers the local
// fallback used when no relay is connected.
func newSquadRoomCode() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("LGVX-%d", time.Now().UnixNano()%900000+100000)
	}
	var sb strings.Builder
	sb.WriteString("LGVX-")
	for _, b := range buf {
		sb.WriteByte(squadRoomCodeAlphabet[int(b)%len(squadRoomCodeAlphabet)])
	}
	return sb.String()
}
