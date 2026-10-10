// Command migrate applies database migrations from the migrations/ directory.
//
// TODO(stage 4): embed goose (github.com/pressly/goose) here, or run the goose
// CLI directly:
//
//	goose -dir migrations postgres "$DATABASE_URL" up
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "migrate: not implemented yet; use the goose CLI (see Makefile target 'migrate-up')")
	os.Exit(2)
}
