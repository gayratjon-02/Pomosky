package config

import "testing"

func TestValidate(t *testing.T) {
	prod := Config{Env: "production", DatabaseURL: "db", HostTokenSalt: "h"}

	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"development needs nothing", Config{Env: "development"}, false},
		{"development may enable dev reload", Config{Env: "development", Dev: true}, false},
		{"complete production config", prod, false},
		{"production without database", Config{Env: "production", HostTokenSalt: "h"}, true},
		{"production without host token salt", Config{Env: "production", DatabaseURL: "db"}, true},
		{"production rejects dev reload", func() Config { c := prod; c.Dev = true; return c }(), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadTrustProxy(t *testing.T) {
	if Load().TrustProxy {
		t.Fatal("TRUST_PROXY must default to false")
	}
	t.Setenv("TRUST_PROXY", "true")
	if !Load().TrustProxy {
		t.Fatal("TRUST_PROXY=true was not picked up")
	}
}

func TestLoadDevFlag(t *testing.T) {
	t.Setenv("DEV_RELOAD", "true")
	t.Setenv("DEV_VIEWS_DIR", "")
	cfg := Load()
	if !cfg.Dev || cfg.DevViewsDir != "internal/views" {
		t.Fatalf("unexpected dev config: %+v", cfg)
	}

	t.Setenv("DEV_RELOAD", "not-a-bool")
	if Load().Dev {
		t.Fatal("an invalid DEV_RELOAD value must fall back to false")
	}
}
