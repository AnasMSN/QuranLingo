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
}

func NewRouter(h *Handlers, authService *service.AuthService) http.Handler {
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

	return r
}
