package google

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func fakeGoogle(t *testing.T, verified bool) *Provider {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("code") != "good-code" || r.PostForm.Get("client_secret") != "secret" || r.PostForm.Get("grant_type") != "authorization_code" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"tok"}`))
	})
	mux.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		v := "false"
		if verified {
			v = "true"
		}
		_, _ = w.Write([]byte(`{"sub":"123","email":"a@example.com","email_verified":` + v + `,"name":"Ayla"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := New("client", "secret", "https://pomosky.test/auth/callback?provider=google")
	p.tokenURL = srv.URL + "/token"
	p.userInfoURL = srv.URL + "/userinfo"
	return p
}

func TestAuthURL(t *testing.T) {
	p := New("client", "secret", "https://pomosky.test/cb")
	u, err := url.Parse(p.AuthURL("st4te"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("state") != "st4te" || q.Get("client_id") != "client" || q.Get("response_type") != "code" || q.Get("redirect_uri") != "https://pomosky.test/cb" {
		t.Fatalf("query = %v", q)
	}
}

func TestExchange(t *testing.T) {
	id, err := fakeGoogle(t, true).Exchange(context.Background(), "good-code")
	if err != nil {
		t.Fatal(err)
	}
	if id.Provider != "google" || id.Subject != "123" || id.Email != "a@example.com" || id.Name != "Ayla" {
		t.Fatalf("identity = %+v", id)
	}
}

func TestExchangeRejectsBadCodeAndUnverifiedEmail(t *testing.T) {
	if _, err := fakeGoogle(t, true).Exchange(context.Background(), "wrong"); err == nil {
		t.Fatal("bad code accepted")
	}
	if _, err := fakeGoogle(t, false).Exchange(context.Background(), "good-code"); err == nil {
		t.Fatal("unverified email accepted")
	}
}
