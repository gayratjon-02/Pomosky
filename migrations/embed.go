// Package migrations embeds the SQL migrations so the server binary can apply them.
package migrations

import "embed"

// FS holds every *.sql migration (goose format).
//
//go:embed *.sql
var FS embed.FS
