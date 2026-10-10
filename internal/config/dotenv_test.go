package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotenvLine(t *testing.T) {
	cases := []struct {
		in, key, value string
		ok             bool
	}{
		{"A=1", "A", "1", true},
		{"  export B = two ", "B", "two", true},
		{`C="has # hash and spaces"`, "C", "has # hash and spaces", true},
		{"D='single'", "D", "single", true},
		{"E=value # comment", "E", "value", true},
		{"F=", "F", "", true},
		{"G=a=b", "G", "a=b", true},
		{"# comment", "", "", false},
		{"", "", "", false},
		{"no equals", "", "", false},
		{"=x", "", "", false},
		{"bad key=1", "", "", false},
	}
	for _, tc := range cases {
		k, v, ok := parseDotenvLine(tc.in)
		if k != tc.key || v != tc.value || ok != tc.ok {
			t.Errorf("%q = (%q, %q, %v), want (%q, %q, %v)", tc.in, k, v, ok, tc.key, tc.value, tc.ok)
		}
	}
}

func TestLoadDotenvDoesNotOverrideTheEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("POMOSKY_T1=from-file\nPOMOSKY_T2=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("POMOSKY_T2", "from-env")
	os.Unsetenv("POMOSKY_T1")
	t.Cleanup(func() { os.Unsetenv("POMOSKY_T1") })

	if err := LoadDotenv(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("POMOSKY_T1") != "from-file" || os.Getenv("POMOSKY_T2") != "from-env" {
		t.Fatalf("T1=%q T2=%q", os.Getenv("POMOSKY_T1"), os.Getenv("POMOSKY_T2"))
	}
	if err := LoadDotenv(filepath.Join(t.TempDir(), "missing")); err != nil {
		t.Fatalf("a missing file must not be an error: %v", err)
	}
}

// Every variable the code reads must be documented, so the docs cannot drift.
func TestEveryEnvVariableIsDocumented(t *testing.T) {
	src, err := os.ReadFile("config.go")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile("../../docs/env.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range envNames(string(src)) {
		if !contains(string(doc), "`"+name+"`") {
			t.Errorf("%s is read in config.go but not documented in docs/env.md", name)
		}
	}
}

func envNames(src string) []string {
	var names []string
	seen := map[string]bool{}
	for _, call := range []string{`getenv("`, `getenvBool("`, `getenvInt("`, `getenvDuration("`, `os.Getenv("`} {
		rest := src
		for {
			i := indexOf(rest, call)
			if i < 0 {
				break
			}
			rest = rest[i+len(call):]
			j := indexOf(rest, `"`)
			if j <= 0 {
				break
			}
			if n := rest[:j]; !seen[n] && n != "" && n == upper(n) {
				seen[n] = true
				names = append(names, n)
			}
		}
	}
	return names
}

func contains(s, sub string) bool { return indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}

func TestSplitEmails(t *testing.T) {
	got := splitEmails(" Boss@Example.com, other@example.com;third@example.com\nfourth@example.com ,, ")
	want := []string{"boss@example.com", "other@example.com", "third@example.com", "fourth@example.com"}
	if len(got) != len(want) {
		t.Fatalf("%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v", got)
		}
	}
	if splitEmails("") != nil {
		t.Fatal("empty input should give nil")
	}
}
