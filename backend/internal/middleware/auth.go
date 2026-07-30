package middleware

import (
	"context"
	"net/http"
	"strings"

	"quranlingo/backend/internal/httpx"
	"quranlingo/backend/internal/service"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// Auth requires a valid "Authorization: Bearer <access token>" header and
// injects the authenticated user ID into the request context.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				httpx.Error(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}

			userID, err := authService.ParseAccessToken(strings.TrimPrefix(header, prefix))
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "invalid or expired access token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDContextKey).(string)
	return v, ok
}
