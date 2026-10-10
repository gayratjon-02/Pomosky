package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const internalPrefix = "github.com/gayratjon-02/Pomosky/internal/"

// allowed lists, for each package, the internal packages it may import.
// Anything not listed is a violation. internal/app (the composition root) and
// the cmd packages are intentionally not restricted.
var allowed = map[string][]string{
	"apperr":             {},
	"config":             {},
	"room":               {},
	"weather":            {},
	"weather/openmeteo":  {"weather", "apperr"},
	"models":             {"room"},
	"ports":              {"room", "models"},
	"platform/clock":     {},
	"platform/logger":    {},
	"platform/idgen":     {"room"},
	"platform/httpx":     {"apperr"},
	"platform/ratelimit": {},
	"platform/reqctx":    {"models"},
	"moderation":         {},
	"mail":               {"ports"},
	"stats":              {"models", "ports", "apperr"},
	"devreload":          {},
	"analytics":          {"room"},
	"views":              {"models"},
	"auth":               {"models", "ports", "apperr"},
	"auth/google":        {"auth", "models", "apperr"},
	"store/postgres":     {"models", "ports", "room", "apperr", "store/postgres/sqlcgen"},
	"realtime":           {"room", "models", "ports", "moderation", "platform/ratelimit"},
	"middleware":         {"models", "platform/httpx", "platform/ratelimit", "platform/reqctx"},
	"controllers":        {"models", "ports", "room", "views", "config", "weather", "platform/httpx", "platform/reqctx", "apperr"},
}

// pure packages may not import any third-party module either.
var pure = map[string]bool{"apperr": true, "room": true, "weather": true, "models": true}

func TestDependencyRules(t *testing.T) {
	for pkg, allow := range allowed {
		dir := filepath.Join("..", filepath.FromSlash(pkg))
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Errorf("rule for package %q but directory is missing: %v", pkg, err)
			continue
		}
		ok := map[string]bool{}
		for _, a := range allow {
			ok[a] = true
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, parser.ImportsOnly)
			if err != nil {
				t.Errorf("parse %s/%s: %v", pkg, name, err)
				continue
			}
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(path, internalPrefix) {
					dep := strings.TrimPrefix(path, internalPrefix)
					if !ok[dep] {
						t.Errorf("internal/%s/%s imports internal/%s, which is not allowed (allowed: %v)", pkg, name, dep, allow)
					}
					continue
				}
				if pure[pkg] {
					first, _, _ := strings.Cut(path, "/")
					if strings.Contains(first, ".") {
						t.Errorf("pure package internal/%s/%s imports third-party %q", pkg, name, path)
					}
				}
			}
		}
	}
}
