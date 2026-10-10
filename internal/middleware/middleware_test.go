package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gayratjon-02/Pomosky/internal/models"
)

func TestChainOrder(t *testing.T) {
	var order []string
	mk := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { order = append(order, "handler") }),
		mk("a"), mk("b"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if got := strings.Join(order, ","); got != "a,b,handler" {
		t.Fatalf("order = %s", got)
	}
}

func TestRecoverReturns500(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }), Recover(log))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

type failingResolver struct{}

func (failingResolver) UserFromSession(context.Context, string) (*models.User, error) {
	return nil, context.Canceled
}

func TestOptionalAuthNeverBlocks(t *testing.T) {
	reached := false
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		if UserFromContext(r.Context()) != nil {
			t.Error("user should be anonymous")
		}
	}), OptionalAuth(failingResolver{}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "pomosky_session", Value: "bad"})
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !reached {
		t.Fatal("OptionalAuth blocked the request")
	}
}
