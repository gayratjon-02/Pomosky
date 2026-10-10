package views

import (
	"html/template"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func funcMap(r *Renderer) template.FuncMap {
	return template.FuncMap{
		"year":    func() int { return time.Now().Year() },
		"initial": initial,
		// asset turns a logical name ("base.css", "/static/landing.js") into the URL of the built file.
		"asset": func(name string) string { return r.resolve(name) },
	}
}

// initial returns the capitalised first letter of a name, "?" for an empty one.
func initial(name string) string {
	name = strings.TrimSpace(name)
	r, size := utf8.DecodeRuneInString(name)
	if size == 0 || r == utf8.RuneError {
		return "?"
	}
	return string(unicode.ToUpper(r))
}
