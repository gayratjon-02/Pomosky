package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed locales/*.json
var files embed.FS

// Lang is a language the site is offered in.
type Lang string

// The supported languages. English is the source language and the fallback.
const (
	EN Lang = "en"
	UZ Lang = "uz"
	RU Lang = "ru"
	KO Lang = "ko"
)

// Supported lists the languages in the order the switcher shows them.
var Supported = []Lang{EN, UZ, RU, KO}

// Cookie is the name of the cookie that remembers the visitor's language.
const Cookie = "pomosky_lang"

// Names are the languages written in themselves, for the switcher.
var Names = map[Lang]string{EN: "English", UZ: "Oʻzbekcha", RU: "Русский", KO: "한국어"}

var dict = map[Lang]map[string]string{}

func init() {
	for _, l := range Supported {
		if l == EN {
			continue
		}
		raw, err := files.ReadFile("locales/" + string(l) + ".json")
		if err != nil {
			panic(fmt.Sprintf("i18n: %v", err))
		}
		m := map[string]string{}
		if err := json.Unmarshal(raw, &m); err != nil {
			panic(fmt.Sprintf("i18n: locales/%s.json: %v", l, err))
		}
		dict[l] = m
	}
}

// Valid reports whether s is a supported language code.
func Valid(s string) bool {
	for _, l := range Supported {
		if string(l) == s {
			return true
		}
	}
	return false
}

// Parse returns the language for a code, English for anything unknown.
func Parse(s string) Lang {
	if Valid(s) {
		return Lang(s)
	}
	return EN
}

// Pick chooses the language of a request: the cookie if it holds a supported
// one, else the best match of the Accept-Language header, else English.
func Pick(cookie, acceptLanguage string) Lang {
	if Valid(cookie) {
		return Lang(cookie)
	}
	type cand struct {
		lang Lang
		q    float64
		pos  int
	}
	var best []cand
	for i, part := range strings.Split(acceptLanguage, ",") {
		tag, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		q := 1.0
		if v, ok := strings.CutPrefix(strings.TrimSpace(params), "q="); ok {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				q = f
			}
		}
		primary, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(tag)), "-")
		if q > 0 && Valid(primary) {
			best = append(best, cand{Lang(primary), q, i})
		}
	}
	if len(best) == 0 {
		return EN
	}
	sort.SliceStable(best, func(i, j int) bool { return best[i].q > best[j].q })
	return best[0].lang
}

// T translates an English text, falling back to the text itself.
func T(l Lang, key string) string {
	if l == EN {
		return key
	}
	if v, ok := dict[l][key]; ok && v != "" {
		return v
	}
	return key
}

// Keys returns the keys a language has translated (for tests).
func Keys(l Lang) []string {
	out := make([]string, 0, len(dict[l]))
	for k := range dict[l] {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
