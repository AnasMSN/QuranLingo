package middleware

import (
	"crypto/subtle"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// AdminBasicAuth gates the operator-only /admin panel behind HTTP Basic Auth,
// checked against a single credential from config — never a regular user
// account, so there's no risk of a normal signup ever becoming an admin.
// The password is compared as a bcrypt hash, same primitive already used for
// user passwords in auth_service.go.
func AdminBasicAuth(username, passwordHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			validUser := ok && subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
			validPass := ok && bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(pass)) == nil

			if !validUser || !validPass {
				w.Header().Set("WWW-Authenticate", `Basic realm="QuranLingo Admin"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
