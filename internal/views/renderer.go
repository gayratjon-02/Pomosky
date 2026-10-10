package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Renderer renders full pages. Each page is parsed together with the layouts
// and partials into its own template set, so pages can define the same blocks.
type Renderer struct {
	fsys   fs.FS
	reload bool // development: re-parse the templates on every request

	pages map[string]*template.Template // parsed once; unused when reload is set

	resolveAsset func(string) string // maps logical asset names to built URLs; nil keeps /static/<name>
}

// SetAssetResolver sets how the asset template function turns a logical name
// into a URL (the app passes the Vite manifest). Call it before serving.
func (r *Renderer) SetAssetResolver(fn func(string) string) { r.resolveAsset = fn }

func (r *Renderer) resolve(name string) string {
	if r.resolveAsset != nil {
		return r.resolveAsset(name)
	}
	return "/static/" + strings.TrimPrefix(name, "/static/")
}

// New parses every templates/pages/*.html from fsys once.
func New(fsys fs.FS) (*Renderer, error) {
	r := &Renderer{fsys: fsys}
	pages, err := parsePages(fsys, r)
	if err != nil {
		return nil, err
	}
	r.pages = pages
	return r, nil
}

// NewReloading is like New but re-parses the templates on every request, so
// edits show up immediately. Template errors are shown in the response body.
// Use it for local development only.
func NewReloading(fsys fs.FS) (*Renderer, error) {
	r, err := New(fsys)
	if err != nil {
		return nil, err
	}
	r.reload = true
	return r, nil
}

// FromDir returns a reloading Renderer that reads the templates from
// dir/templates on disk.
func FromDir(dir string) (*Renderer, error) {
	if _, err := os.Stat(filepath.Join(dir, "templates")); err != nil {
		return nil, fmt.Errorf("views: %s/templates not found (run from the repository root or set DEV_VIEWS_DIR): %w", dir, err)
	}
	return NewReloading(os.DirFS(dir))
}

func parsePages(fsys fs.FS, r *Renderer) (map[string]*template.Template, error) {
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
		t, err := template.New(name).Funcs(funcMap(r)).
			ParseFS(fsys, "templates/layouts/*.html", "templates/partials/*.html", file)
		if err != nil {
			return nil, fmt.Errorf("views: parse %s: %w", file, err)
		}
		pages[name] = t
	}
	return pages, nil
}

func (r *Renderer) lookup(page string) (*template.Template, error) {
	pages := r.pages
	if r.reload {
		var err error
		if pages, err = parsePages(r.fsys, r); err != nil {
			return nil, err
		}
	}
	t, ok := pages[page]
	if !ok {
		return nil, fmt.Errorf("views: page template not found: %s", page)
	}
	return t, nil
}

// Render executes the named page with the status code. The page is rendered
// into a buffer first so a template error never produces a half-written page.
func (r *Renderer) Render(w http.ResponseWriter, status int, page string, data PageData) {
	t, err := r.lookup(page)
	if err == nil {
		var buf bytes.Buffer
		if err = t.ExecuteTemplate(&buf, "base", data); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(status)
			_, _ = buf.WriteTo(w)
			return
		}
	}
	msg := "internal server error"
	if r.reload {
		msg = err.Error() // development: show the template error in the browser
	}
	http.Error(w, msg, http.StatusInternalServerError)
}

// Has reports whether a page template exists.
func (r *Renderer) Has(page string) bool {
	if r.reload {
		_, err := fs.Stat(r.fsys, path.Join("templates", "pages", page+".html"))
		return err == nil
	}
	_, ok := r.pages[page]
	return ok
}
