// Package idgen generates unguessable identifiers with crypto/rand.
package idgen

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"strings"

	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Generator implements ports.IDGenerator.
type Generator struct{}

var roomEnc = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewRoomID returns an 8-character lowercase base32 ID.
func (Generator) NewRoomID() room.ID {
	b := make([]byte, 5)
	mustRead(b)
	return room.ID(strings.ToLower(roomEnc.EncodeToString(b)))
}

// NewToken returns a 256-bit URL-safe random token.
func (Generator) NewToken() string {
	b := make([]byte, 32)
	mustRead(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// NewSessionID returns a new session identifier.
func (g Generator) NewSessionID() string { return g.NewToken() }

func mustRead(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("idgen: no randomness available: " + err.Error())
	}
}
