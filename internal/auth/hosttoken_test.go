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

func TestHostTokensType(t *testing.T) {
	tokens := HostTokens{Salt: "salt"}
	hash := tokens.Hash("secret")
	if hash != HashHostToken("salt", "secret") {
		t.Fatal("Hash must match HashHostToken")
	}
	if !tokens.Verify("secret", hash) || tokens.Verify("other", hash) {
		t.Fatal("Verify gave the wrong answer")
	}
	if tokens.Verify("", hash) {
		t.Fatal("an empty token must never verify")
	}
}
