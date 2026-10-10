// Package assets maps logical names of built frontend files ("landing.js",
// "base.css") to the hashed file names Vite writes ("landing.CWdZvdpn.js").
//
// Hashed names change whenever the content changes, so a browser can never
// combine an old entry file with new chunks, and hashed files can be cached for
// a year. Without a manifest (for example before the first build) names resolve
// to the plain /static/<name> path, so the site still works.
package assets

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	urlPrefix    = "/static/"
	manifestPath = ".vite/manifest.json"
)

// Manifest resolves logical asset names.
type Manifest struct {
	dir   string
	watch bool // re-read the manifest when it changes (development)

	mu      sync.Mutex
	files   map[string]string // logical name -> file relative to dir
	modTime time.Time
}

// Load reads dir/.vite/manifest.json. A missing or broken manifest is not an
// error: Resolve then falls back to the plain names. With watch set the file is
// read again whenever it changes, which suits `vite build --watch`.
func Load(dir string, watch bool) *Manifest {
	m := &Manifest{dir: dir, watch: watch}
	m.refresh()
	return m
}

// Resolve returns the URL of a built file. name may be "landing.js", "base.css"
// or already a URL such as "/static/landing.js"; anything the manifest does not
// know is returned as /static/<name>.
func (m *Manifest) Resolve(name string) string {
	name = strings.TrimPrefix(name, urlPrefix)
	if m == nil {
		return urlPrefix + name
	}
	if m.watch {
		m.refresh()
	}
	m.mu.Lock()
	file, ok := m.files[name]
	m.mu.Unlock()
	if !ok {
		return urlPrefix + name
	}
	return urlPrefix + file
}

func (m *Manifest) refresh() {
	full := filepath.Join(m.dir, manifestPath)
	info, err := os.Stat(full)
	if err != nil {
		m.mu.Lock()
		m.files, m.modTime = nil, time.Time{}
		m.mu.Unlock()
		return
	}
	m.mu.Lock()
	unchanged := info.ModTime().Equal(m.modTime) && m.files != nil
	m.mu.Unlock()
	if unchanged {
		return
	}
	raw, err := os.ReadFile(full)
	if err != nil {
		return
	}
	files := parse(raw)
	m.mu.Lock()
	m.files, m.modTime = files, info.ModTime()
	m.mu.Unlock()
}

// parse reads the Vite manifest and keeps the entry files, keyed by "<name><ext>".
func parse(raw []byte) map[string]string {
	var entries map[string]struct {
		File    string `json:"file"`
		Name    string `json:"name"`
		IsEntry bool   `json:"isEntry"`
	}
	if json.Unmarshal(raw, &entries) != nil {
		return map[string]string{}
	}
	files := make(map[string]string, len(entries))
	for _, e := range entries {
		if !e.IsEntry || e.File == "" || e.Name == "" {
			continue
		}
		files[e.Name+path.Ext(e.File)] = e.File
	}
	return files
}

// Hashed reports whether a built file name carries a content hash, like
// "landing.CWdZvdpn.js" or "chunks/astro.BY394th4.js". Such files never change.
func Hashed(file string) bool {
	base := path.Base(file)
	parts := strings.Split(base, ".")
	if len(parts) < 3 {
		return false
	}
	hash := parts[len(parts)-2]
	if len(hash) < 6 {
		return false
	}
	for _, r := range hash {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
