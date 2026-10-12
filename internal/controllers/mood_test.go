package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func homeWithMood(t *testing.T, value string) string {
	t.Helper()
	c := newTestControllers(t)
	req := httptest.NewRequest("GET", "/", nil)
	if value != "" {
		req.AddCookie(&http.Cookie{Name: "pomosky_mood", Value: value})
	}
	rec := httptest.NewRecorder()
	c.Home(rec, req)
	return rec.Body.String()
}

func TestPageIsRenderedInTheChosenMood(t *testing.T) {
	if body := homeWithMood(t, "sunset"); !strings.Contains(body, `<html lang="en" data-mood="sunset">`) {
		t.Fatalf("the mood is missing from the page: %.200s", body)
	}
}

func TestUnknownOrLiveMoodIsIgnored(t *testing.T) {
	for _, value := range []string{"", "live", "nope", `"><script>`} {
		if body := homeWithMood(t, value); strings.Contains(body, "data-mood") {
			t.Errorf("cookie %q must not reach the page", value)
		}
	}
}
