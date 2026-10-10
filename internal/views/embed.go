package views

import "embed"

// FS contains all templates, embedded into the binary.
//
//go:embed templates
var FS embed.FS

// Default builds a Renderer from the embedded templates.
func Default() (*Renderer, error) { return New(FS) }
