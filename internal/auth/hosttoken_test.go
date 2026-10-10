package auth

import "testing"

func TestHostToken(t *testing.T) {
	h := HashHostToken("salt", "token-1")
	if !VerifyHostToken("salt", "token-1", h) {
		t.Fatal("valid token rejected")
	}
	if VerifyHostToken("salt", "token-2", h) {
		t.Fatal("wrong token accepted")
	}
	if VerifyHostToken("other-salt", "token-1", h) {
		t.Fatal("wrong salt accepted")
	}
}
