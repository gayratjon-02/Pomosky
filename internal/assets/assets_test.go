package assets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sample = `{
  "src/entries/landing.ts": {"file": "landing.CWdZvdpn.js", "name": "landing", "isEntry": true, "imports": ["_astro.BY394th4.js"]},
  "src/styles/base.css": {"file": "base.QFNHrnKh.css", "name": "base", "isEntry": true},
  "_astro.BY394th4.js": {"file": "chunks/astro.BY394th4.js", "name": "astro"}
}`

func writeManifest(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".vite"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifestPath), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveUsesTheManifest(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, sample)
	m := Load(dir, false)

	cases := map[string]string{
		"landing.js":         "/static/landing.CWdZvdpn.js",
		"/static/landing.js": "/static/landing.CWdZvdpn.js",
		"base.css":           "/static/base.QFNHrnKh.css",
		"favicon.svg":        "/static/favicon.svg", // not built by Vite: unchanged
		"astro.js":           "/static/astro.js",    // a chunk is not an entry
	}
	for in, want := range cases {
		if got := m.Resolve(in); got != want {
			t.Errorf("Resolve(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveFallsBackWithoutAManifest(t *testing.T) {
	if got := Load(t.TempDir(), false).Resolve("landing.js"); got != "/static/landing.js" {
		t.Fatalf("got %q", got)
	}
	var none *Manifest
	if got := none.Resolve("/static/base.css"); got != "/static/base.css" {
		t.Fatalf("nil manifest: %q", got)
	}
	dir := t.TempDir()
	writeManifest(t, dir, "not json")
	if got := Load(dir, false).Resolve("landing.js"); got != "/static/landing.js" {
		t.Fatalf("broken manifest: %q", got)
	}
}

func TestWatchPicksUpANewBuild(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, sample)
	m := Load(dir, true)
	if got := m.Resolve("landing.js"); got != "/static/landing.CWdZvdpn.js" {
		t.Fatalf("first build: %q", got)
	}
	writeManifest(t, dir, `{"src/entries/landing.ts": {"file": "landing.NEWHASH1.js", "name": "landing", "isEntry": true}}`)
	later := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(dir, manifestPath), later, later); err != nil {
		t.Fatal(err)
	}
	if got := m.Resolve("landing.js"); got != "/static/landing.NEWHASH1.js" {
		t.Fatalf("second build: %q", got)
	}
}

func TestHashed(t *testing.T) {
	for file, want := range map[string]bool{
		"landing.CWdZvdpn.js":      true,
		"chunks/astro.BY394th4.js": true,
		"base.QFNHrnKh.css":        true,
		"landing.js":               false,
		"favicon.svg":              false,
		"og.png":                   false,
		"chunks/astro.js":          false,
		"landing.short.js":         false, // hash shorter than 6 characters
		"weird name.with space.js": false,
	} {
		if got := Hashed(file); got != want {
			t.Errorf("Hashed(%q) = %v, want %v", file, got, want)
		}
	}
}
