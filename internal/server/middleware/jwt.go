package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ContextUserID    contextKey = "user_id"
	ContextUserEmail contextKey = "user_email"
)

// ParseTokenFn is a function that validates a raw JWT and returns claims.
type ParseTokenFn func(token string) (userID, email string, err error)

func JWT(parse ParseTokenFn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := tokenFromRequest(r)
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, email, err := parse(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextUserEmail, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ContextUserID).(string)
	return v
}

func tokenFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
