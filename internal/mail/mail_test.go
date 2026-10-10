package mail

import (
	"strings"
	"testing"
	"time"
)

func TestBuildMessage(t *testing.T) {
	msg, err := buildMessage("Pomosky <no-reply@pomosky.com>", "a@example.com", "Hello", "line1\nline2", time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	s := string(msg)
	for _, want := range []string{"To: a@example.com\r\n", "Subject: Hello\r\n", "\r\n\r\nline1\r\nline2"} {
		if !strings.Contains(s, want) {
			t.Errorf("message lacks %q:\n%s", want, s)
		}
	}
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	if _, err := buildMessage("a@b.c", "x@example.com\r\nBcc: evil@example.com", "s", "b", time.Now()); err == nil {
		t.Fatal("header injection accepted")
	}
	if _, err := buildMessage("a@b.c", "x@example.com", "s\nBcc: evil", "b", time.Now()); err == nil {
		t.Fatal("subject injection accepted")
	}
}
