package views

import (
	"html/template"
	"time"
)

func funcMap() template.FuncMap {
	return template.FuncMap{
		"year": func() int { return time.Now().Year() },
	}
}
