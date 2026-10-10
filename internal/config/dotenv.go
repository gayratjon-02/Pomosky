package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotenv reads KEY=VALUE lines from a file into the process environment.
// Variables that are already set win, so a real environment (Docker, systemd)
// always overrides the file. A missing file is not an error. Use it for local
// development only: production gets its environment from the deployment.
func LoadDotenv(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, value, ok := parseDotenvLine(sc.Text())
		if !ok {
			continue
		}
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

// parseDotenvLine understands "KEY=value", "export KEY=value", quoted values
// and full-line comments. Anything else is skipped.
func parseDotenvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")
	k, v, found := strings.Cut(line, "=")
	k = strings.TrimSpace(k)
	if !found || k == "" || strings.ContainsAny(k, " \t") {
		return "", "", false
	}
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return k, v[1 : len(v)-1], true
	}
	if i := strings.Index(v, " #"); i >= 0 { // trailing comment
		v = strings.TrimSpace(v[:i])
	}
	return k, v, true
}
