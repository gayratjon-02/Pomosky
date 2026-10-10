package views

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedTemplatesRender(t *testing.T) {
	r, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range []string{"home", "room", "host", "record", "login", "account", "privacy", "error"} {
		if !r.Has(page) {
			t.Errorf("missing page template %q", page)
			continue
		}
		rec := httptest.NewRecorder()
		r.Render(rec, http.StatusOK, page, PageData{Title: "Test", Data: RoomPage{ID: "abc", ShareURL: "http://x/r/abc"}})
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", page, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "<title>") {
			t.Errorf("%s: no <title> in output", page)
		}
	}
}
