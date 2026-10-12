package i18n

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// usedKeys finds every text the site asks to have translated: {{$.T "..."}} and
// {{$.H "..."}} in the templates, t("..."), tn("...") in the frontend, and the few
// texts that are held in tables or sent by the server (listed below).
func usedKeys(t *testing.T) map[string]bool {
	t.Helper()
	keys := map[string]bool{}
	unescape := strings.NewReplacer(`\"`, `"`, `\\`, `\`)
	add := func(raw string) { keys[unescape.Replace(raw)] = true }

	tmpl := regexp.MustCompile(`\$\.[TH] "((?:[^"\\]|\\.)*)"`)
	front := regexp.MustCompile(`(?:^|[^\w.])(?:t|tn|tr)\("((?:[^"\\]|\\.)*)"`)
	mood := regexp.MustCompile(`\b(?:name|hint): "((?:[^"\\]|\\.)*)"`)

	walk := func(root string, re *regexp.Regexp, accept func(string) bool) {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !accept(path) {
				return err
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, m := range re.FindAllStringSubmatch(string(src), -1) {
				for _, g := range m[1:] {
					if g != "" {
						add(g)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	walk("../views/templates", tmpl, func(p string) bool { return strings.HasSuffix(p, ".html") })
	walk("../../web/src", front, func(p string) bool {
		return strings.HasSuffix(p, ".ts") && !strings.HasSuffix(p, ".test.ts")
	})
	walk("../../web/src/sky", mood, func(p string) bool { return strings.HasSuffix(p, "moods.ts") })

	// Texts written in Go: page titles and descriptions, error messages and sign-in notices.
	// (Admin and development pages are not translated.)
	goText := regexp.MustCompile(`(?:Title|Description|Data):\s+"((?:[^"\\]|\\.)*)"|title :?= "((?:[^"\\]|\\.)*)"|loginPage\(w, r, [^,]+, "((?:[^"\\]|\\.)*)"`)
	untranslated := map[string]bool{"Admin": true, "Sky lab": true, "Mobile preview": true}
	walk("../controllers", goText, func(p string) bool {
		return strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go")
	})
	for k := range untranslated {
		delete(keys, k)
	}

	for _, k := range []string{
		// held in tables (timer view, connection chip, host controls)
		"Waiting for the host", "Focus", "Break", "Paused", "Connecting", "Live", "Reconnecting", "Disconnected", "Resume", "Pause",
		// sent by the server (push alerts)
		"Break is over", "Ready for the next focus session?", "Focus session finished", "{label}. Time for a break.", "Time for a break.",
		"Language",
	} {
		keys[k] = true
	}
	return keys
}

var placeholder = regexp.MustCompile(`\{\w+\}`)

func placeholders(s string) string {
	seen := map[string]bool{}
	var found []string
	for _, p := range placeholder.FindAllString(s, -1) {
		if !seen[p] {
			seen[p] = true
			found = append(found, p)
		}
	}
	sort.Strings(found)
	return strings.Join(found, ",")
}

func TestEveryTextIsTranslatedIntoEveryLanguage(t *testing.T) {
	used := usedKeys(t)
	if len(used) < 250 {
		t.Fatalf("only %d texts found, the scan is broken", len(used))
	}
	for _, lang := range []Lang{UZ, RU, KO} {
		for key := range used {
			if strings.TrimSpace(dict[lang][key]) == "" {
				t.Errorf("%s: missing translation of %q", lang, key)
			}
		}
		for key := range dict[lang] {
			if !used[key] {
				t.Errorf("%s: %q is not used anywhere (remove it or fix the key)", lang, key)
			}
		}
	}
}

func TestTranslationsKeepPlaceholdersMarkupAndPluralForms(t *testing.T) {
	for _, lang := range []Lang{UZ, RU, KO} {
		for key, tr := range dict[lang] {
			if placeholders(key) != placeholders(tr) {
				t.Errorf("%s: %q has placeholders %q but its translation has %q", lang, key, placeholders(key), placeholders(tr))
			}
			if strings.Count(key, "<") != strings.Count(tr, "<") || strings.Count(key, ">") != strings.Count(tr, ">") {
				t.Errorf("%s: markup of %q changed", lang, key)
			}
			if strings.Contains(key, "<strong>delete</strong>") && !strings.Contains(tr, "<strong>delete</strong>") {
				t.Errorf("%s: the word the person must type (delete) was translated in %q", lang, key)
			}
			if strings.Contains(key, "|") && lang == RU && strings.Count(tr, "|") != 2 {
				t.Errorf("ru: %q needs three plural forms, got %q", key, tr)
			}
		}
	}
}
