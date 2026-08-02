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

// AdminAuth holds the operator credential and session-signing secret the
// /admin panel is gated behind. If any field is empty, the admin panel is
// not mounted at all.
type AdminAuth struct {
	Username      string
	PasswordHash  string
	SessionSecret string
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

	// Operator-only admin panel — its own login page + signed session cookie,
	// entirely separate from regular user JWT auth (no admin registration
	// flow exists; the only admin is whatever ADMIN_USERNAME/ADMIN_PASSWORD_HASH
	// resolve to). Not mounted at all unless username, password hash, and
	// session secret are all set.
	if admin.Username != "" && admin.PasswordHash != "" && admin.SessionSecret != "" {
		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(10, time.Minute))
			r.Get("/admin/login", h.Admin.LoginPage)
			r.Post("/admin/login", h.Admin.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AdminSessionAuth(admin.SessionSecret, admin.Username))
			r.Use(httprate.LimitByIP(30, time.Minute))
			r.Get("/admin", h.Admin.Dashboard)
			r.Post("/admin/logout", h.Admin.Logout)
			r.Post("/admin/users/refill-all", h.Admin.RefillAll)
			r.Post("/admin/users/{id}/refill", h.Admin.RefillOne)

			r.Get("/admin/questions", h.Admin.QuestionsPage)
			r.Post("/admin/questions/new", h.Admin.CreateQuestion)
			r.Post("/admin/questions/import", h.Admin.ImportQuestions)
			r.Post("/admin/questions/{id}/update", h.Admin.UpdateQuestion)
			r.Post("/admin/questions/{id}/approve", h.Admin.ApproveQuestion)
			r.Post("/admin/questions/{id}/delete", h.Admin.DeleteQuestion)
		})
	}

	return r
}
