package views

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyTemplates writes the embedded templates into dir/templates.
func copyTemplates(t *testing.T, dir string) {
	t.Helper()
	err := fs.WalkDir(FS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := fs.ReadFile(FS, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func renderHome(r *Renderer) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.Render(rec, http.StatusOK, "home", PageData{Title: "Test"})
	return rec
}

func writeHome(t *testing.T, dir, content string) {
	t.Helper()
	p := filepath.Join(dir, "templates", "pages", "home.html")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReloadingRendererPicksUpEdits(t *testing.T) {
	dir := t.TempDir()
	copyTemplates(t, dir)

	dev, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	static, err := New(os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}

	writeHome(t, dir, `{{define "content"}}CHANGED-MARKER{{end}}`)

	if body := renderHome(dev).Body.String(); !strings.Contains(body, "CHANGED-MARKER") {
		t.Fatal("reloading renderer did not pick up the edit")
	}
	if body := renderHome(static).Body.String(); strings.Contains(body, "CHANGED-MARKER") {
		t.Fatal("static renderer must keep the templates it parsed at startup")
	}
}

func TestReloadingRendererShowsTemplateErrors(t *testing.T) {
	dir := t.TempDir()
	copyTemplates(t, dir)
	dev, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	writeHome(t, dir, `{{define "content"}}{{ .Broken `)

	rec := renderHome(dev)
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "views:") {
		t.Fatalf("expected the template error in the response, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestStaticRendererHidesErrors(t *testing.T) {
	r, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	r.Render(rec, http.StatusOK, "no-such-page", PageData{})
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "views:") {
		t.Fatalf("static renderer leaked details: %d %q", rec.Code, rec.Body.String())
	}
}

func TestFromDirMissing(t *testing.T) {
	if _, err := FromDir(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected an error for a missing templates directory")
	}
}
