package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashHostToken returns the value stored in the database for a host token.
//
// Host tokens are 256-bit random values, so a salted SHA-256 is sufficient;
// password-style key stretching is not needed. The raw token is shown to the
// host once and never stored.
func HashHostToken(salt, token string) string {
	sum := sha256.Sum256([]byte(salt + ":" + token))
	return hex.EncodeToString(sum[:])
}

// VerifyHostToken reports whether token matches the stored hash, in constant time.
func VerifyHostToken(salt, token, wantHash string) bool {
	got := HashHostToken(salt, token)
	return subtle.ConstantTimeCompare([]byte(got), []byte(wantHash)) == 1
}
