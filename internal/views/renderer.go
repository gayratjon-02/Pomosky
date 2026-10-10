package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Renderer renders full pages. Each page is parsed together with the layouts
// and partials into its own template set, so pages can define the same blocks.
type Renderer struct {
	pages map[string]*template.Template
}

// New parses every templates/pages/*.html from fsys.
func New(fsys fs.FS) (*Renderer, error) {
	files, err := fs.Glob(fsys, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("views: no page templates found")
	}
	pages := make(map[string]*template.Template, len(files))
	for _, file := range files {
		name := strings.TrimSuffix(path.Base(file), ".html")
		t, err := template.New(name).Funcs(funcMap()).
			ParseFS(fsys, "templates/layouts/*.html", "templates/partials/*.html", file)
		if err != nil {
			return nil, fmt.Errorf("views: parse %s: %w", file, err)
		}
		pages[name] = t
	}
	return &Renderer{pages: pages}, nil
}

// Render executes the named page with the status code. The page is rendered
// into a buffer first so a template error never produces a half-written page.
func (r *Renderer) Render(w http.ResponseWriter, status int, page string, data PageData) {
	t, ok := r.pages[page]
	if !ok {
		http.Error(w, "page template not found: "+page, http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// Has reports whether a page template exists.
func (r *Renderer) Has(page string) bool {
	_, ok := r.pages[page]
	return ok
}
