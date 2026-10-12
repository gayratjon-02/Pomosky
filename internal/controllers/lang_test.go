package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func homeIn(t *testing.T, cookie, accept string) string {
	t.Helper()
	c := newTestControllers(t)
	req := httptest.NewRequest("GET", "/", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "pomosky_lang", Value: cookie})
	}
	if accept != "" {
		req.Header.Set("Accept-Language", accept)
	}
	rec := httptest.NewRecorder()
	c.Home(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	return rec.Body.String()
}

func TestPagesAreRenderedInTheVisitorsLanguage(t *testing.T) {
	cases := []struct {
		cookie, accept, lang, text string
	}{
		{"", "", "en", "Start a room"},
		{"uz", "", "uz", "Xona ochish"},
		{"ru", "ko", "ru", "Создать комнату"}, // the cookie wins over the browser
		{"", "ko-KR,ko;q=0.9", "ko", "방 만들기"},
		{"xx", "uz", "uz", "Xona ochish"}, // unknown cookie: the browser decides
	}
	for _, c := range cases {
		body := homeIn(t, c.cookie, c.accept)
		if !strings.Contains(body, `<html lang="`+c.lang+`"`) || !strings.Contains(body, c.text) {
			t.Errorf("cookie %q, accept %q: want lang %s and %q", c.cookie, c.accept, c.lang, c.text)
		}
	}
}

func TestEveryPageRendersInEveryLanguage(t *testing.T) {
	c := newTestControllers(t)
	for _, lang := range []string{"en", "uz", "ru", "ko"} {
		req := httptest.NewRequest("GET", "/privacy", nil)
		req.AddCookie(&http.Cookie{Name: "pomosky_lang", Value: lang})
		rec := httptest.NewRecorder()
		c.Privacy(rec, req)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `<html lang="`+lang+`"`) {
			t.Errorf("privacy in %s: status %d", lang, rec.Code)
		}
		if lang != "en" && strings.Contains(rec.Body.String(), "Pomosky is built to need as little") {
			t.Errorf("privacy in %s was left in English", lang)
		}
	}
}

func TestTheLanguageSwitcherListsAllLanguages(t *testing.T) {
	body := homeIn(t, "ko", "")
	for _, want := range []string{`data-lang="en"`, `data-lang="uz"`, `data-lang="ru"`, `data-lang="ko"`, "Oʻzbekcha", "Русский", "한국어"} {
		if !strings.Contains(body, want) {
			t.Errorf("switcher misses %q", want)
		}
	}
}
