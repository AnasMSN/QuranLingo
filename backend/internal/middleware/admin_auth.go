package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminSessionCookie is the name of the signed session cookie issued on
// successful admin login. It carries no secrets itself (bcrypt hash/password
// never leave the server) — only the admin username and an expiry, HMAC-signed
// so it can't be forged or tampered with without ADMIN_SESSION_SECRET.
const AdminSessionCookie = "admin_session"

const adminSessionTTL = 12 * time.Hour

// VerifyAdminCredentials checks a submitted username/password against the
// single operator credential configured via ADMIN_USERNAME/ADMIN_PASSWORD_HASH.
// There is deliberately no code path that creates or registers an admin
// account — the only admin is whatever is baked into the environment.
func VerifyAdminCredentials(username, passwordHash, gotUser, gotPass string) bool {
	validUser := subtle.ConstantTimeCompare([]byte(gotUser), []byte(username)) == 1
	validPass := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(gotPass)) == nil
	return validUser && validPass
}

// NewAdminSessionCookie builds a signed, HttpOnly session cookie for the given
// admin username. secure should be true whenever the app is served over TLS
// (production) — set false for plain-HTTP local development, otherwise
// browsers silently drop the cookie.
func NewAdminSessionCookie(sessionSecret, username string, secure bool) *http.Cookie {
	expiry := time.Now().Add(adminSessionTTL).Unix()
	token := signAdminSession(sessionSecret, username, expiry)
	return &http.Cookie{
		Name:     AdminSessionCookie,
		Value:    token,
		Path:     "/admin",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(expiry, 0),
	}
}

// ExpiredAdminSessionCookie clears the session cookie on logout.
func ExpiredAdminSessionCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     AdminSessionCookie,
		Value:    "",
		Path:     "/admin",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
}

func signAdminSession(secret, username string, expiryUnix int64) string {
	payload := fmt.Sprintf("%s|%d", username, expiryUnix)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + sig
}

func verifyAdminSession(secret, expectedUsername, token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	encodedPayload, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return false
	}
	payloadParts := strings.SplitN(string(payloadBytes), "|", 2)
	if len(payloadParts) != 2 {
		return false
	}
	username, expiryStr := payloadParts[0], payloadParts[1]

	expiryUnix, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil || time.Now().Unix() > expiryUnix {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(username), []byte(expectedUsername)) == 1
}

// AdminSessionAuth gates the operator-only /admin panel behind a signed
// session cookie issued by the dedicated /admin/login form, instead of the
// browser's native HTTP Basic Auth prompt. Unauthenticated/expired/tampered
// requests are redirected to the login page rather than getting a 401.
func AdminSessionAuth(sessionSecret, username string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(AdminSessionCookie)
			if err != nil || !verifyAdminSession(sessionSecret, username, cookie.Value) {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
