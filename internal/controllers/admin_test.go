package controllers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gayratjon-02/Pomosky/internal/config"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/platform/reqctx"
)

type fakeAdmin struct{ calls int }

func (f *fakeAdmin) Overview(context.Context, string) (models.AdminOverview, error) {
	f.calls++
	return models.AdminOverview{Live: models.LiveSummary{Rooms: 2}, Totals: models.AdminTotals{Users: 3}}, nil
}
func (f *fakeAdmin) Users(context.Context, bool, int, int) ([]models.AdminUser, int, error) {
	f.calls++
	return []models.AdminUser{{Email: "a@example.com", Nickname: "Ayla", Provider: "google"}}, 1, nil
}

func adminSetup(t *testing.T, admins ...string) (*Controllers, *fakeAdmin) {
	t.Helper()
	c := newTestControllers(t)
	c.deps.Config = config.Config{AdminEmails: admins}
	c.deps.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	fa := &fakeAdmin{}
	c.deps.Admin = fa
	return c, fa
}

func asUser(path, email string) *http.Request {
	req := httptest.NewRequest("GET", path, nil)
	if email == "" {
		return req
	}
	e := email
	return req.WithContext(reqctx.WithUser(req.Context(), &models.User{ID: "u1", Email: &e, Nickname: "X"}))
}

func TestAdminIsHiddenFromEveryoneButOperators(t *testing.T) {
	c, fa := adminSetup(t, "boss@example.com")
	for _, tc := range []struct {
		name, email string
		want        int
	}{
		{"anonymous", "", http.StatusNotFound},
		{"another user", "other@example.com", http.StatusNotFound},
		{"operator", "boss@example.com", http.StatusOK},
		{"operator, other letter case", "Boss@Example.com", http.StatusOK},
	} {
		page, api, users := httptest.NewRecorder(), httptest.NewRecorder(), httptest.NewRecorder()
		c.Admin(page, asUser("/admin", tc.email))
		c.AdminOverviewAPI(api, asUser("/api/admin/overview", tc.email))
		c.AdminUsersAPI(users, asUser("/api/admin/users", tc.email))
		for name, rec := range map[string]*httptest.ResponseRecorder{"page": page, "overview": api, "users": users} {
			if rec.Code != tc.want {
				t.Errorf("%s / %s: status %d, want %d", tc.name, name, rec.Code, tc.want)
			}
		}
		if tc.want == http.StatusNotFound && strings.Contains(page.Body.String(), "Operator view") {
			t.Errorf("%s: the admin page leaked", tc.name)
		}
	}
	if fa.calls != 4 { // only the two operator rounds reached the data (overview + users each)
		t.Fatalf("data was read %d times", fa.calls)
	}
}

func TestAdminNeedsAConfiguredList(t *testing.T) {
	c, _ := adminSetup(t) // ADMIN_EMAILS empty: nobody is an operator
	rec := httptest.NewRecorder()
	c.AdminOverviewAPI(rec, asUser("/api/admin/overview", "boss@example.com"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAdminResponsesAreNotCachedOrIndexed(t *testing.T) {
	c, _ := adminSetup(t, "boss@example.com")
	rec := httptest.NewRecorder()
	c.AdminOverviewAPI(rec, asUser("/api/admin/overview", "boss@example.com"))
	if rec.Header().Get("Cache-Control") != "no-store" || !strings.Contains(rec.Header().Get("X-Robots-Tag"), "noindex") {
		t.Fatalf("headers: %v", rec.Header())
	}
	if !strings.Contains(rec.Body.String(), `"users":3`) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestAdminOnlyOpensOnTheAdminHost(t *testing.T) {
	c, _ := adminSetup(t, "boss@example.com")
	c.deps.Config.AdminHost = "dash.example.com"
	for _, tc := range []struct {
		host string
		want int
	}{
		{"dash.example.com", http.StatusOK},
		{"DASH.example.com:443", http.StatusOK},
		{"example.com", http.StatusNotFound},
		{"www.example.com", http.StatusNotFound},
	} {
		req := asUser("/admin", "boss@example.com")
		req.Host = tc.host
		rec, api := httptest.NewRecorder(), httptest.NewRecorder()
		c.Admin(rec, req)
		apiReq := asUser("/api/admin/overview", "boss@example.com")
		apiReq.Host = tc.host
		c.AdminOverviewAPI(api, apiReq)
		if rec.Code != tc.want || api.Code != tc.want {
			t.Errorf("host %q: page %d, api %d, want %d", tc.host, rec.Code, api.Code, tc.want)
		}
	}
}

func TestAdminHostRootGoesToTheAdminPage(t *testing.T) {
	c, _ := adminSetup(t, "boss@example.com")
	c.deps.Config.AdminHost = "dash.example.com"
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "dash.example.com"
	rec := httptest.NewRecorder()
	c.Home(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin" {
		t.Fatalf("got %d %q", rec.Code, rec.Header().Get("Location"))
	}
	req.Host = "example.com"
	rec = httptest.NewRecorder()
	c.Home(rec, req)
	if rec.Code == http.StatusFound {
		t.Fatal("the normal host must keep its landing page")
	}
}
