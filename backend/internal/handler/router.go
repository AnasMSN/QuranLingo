package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"quranlingo/backend/internal/middleware"
	"quranlingo/backend/internal/service"
)

type Handlers struct {
	Auth        *AuthHandler
	Content     *ContentHandler
	Lesson      *LessonHandler
	Leaderboard *LeaderboardHandler
	Admin       *AdminHandler
}

// AdminAuth holds the operator credential the /admin panel is gated behind.
// Both fields empty means the admin panel is not mounted at all.
type AdminAuth struct {
	Username     string
	PasswordHash string
}

func NewRouter(h *Handlers, authService *service.AuthService, admin AdminAuth) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", Health)

	// Auth endpoints are rate-limited per-IP to blunt credential-stuffing/brute force.
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, time.Minute))
		r.Post("/auth/register", h.Auth.Register)
		r.Post("/auth/login", h.Auth.Login)
		r.Post("/auth/refresh", h.Auth.RefreshToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService))

		r.Get("/me", h.Auth.Me)
		r.Get("/courses/{code}", h.Content.GetCourse)
		r.Get("/lessons/{id}", h.Content.GetLesson)
		r.Get("/leaderboard/weekly", h.Leaderboard.Weekly)

		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(30, time.Minute))
			r.Post("/lessons/{id}/submit", h.Lesson.Submit)
		})
	})

	// Operator-only admin panel — HTTP Basic Auth against a config-supplied
	// credential, entirely separate from regular user JWT auth. Not mounted
	// at all unless both ADMIN_USERNAME and ADMIN_PASSWORD_HASH are set.
	if admin.Username != "" && admin.PasswordHash != "" {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AdminBasicAuth(admin.Username, admin.PasswordHash))
			r.Use(httprate.LimitByIP(30, time.Minute))
			r.Get("/admin", h.Admin.Dashboard)
			r.Post("/admin/users/refill-all", h.Admin.RefillAll)
			r.Post("/admin/users/{id}/refill", h.Admin.RefillOne)
		})
	}

	return r
}
