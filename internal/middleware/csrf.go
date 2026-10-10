package middleware

import "net/http"

// CSRF protects state-changing form posts.
//
// TODO(stage 4): implement the double-submit cookie pattern (token in a cookie
// and in a hidden form field, compared on POST). Until then it is a no-op.
func CSRF() Middleware {
	return func(next http.Handler) http.Handler { return next }
}
