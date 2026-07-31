package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"quranlingo/backend/internal/config"
	"quranlingo/backend/internal/db"
	"quranlingo/backend/internal/handler"
	"quranlingo/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	authService := service.NewAuthService(pool, cfg)
	contentService := service.NewContentService(pool)
	lessonService := service.NewLessonService(pool)
	leaderboardService := service.NewLeaderboardService(pool)
	adminService := service.NewAdminService(pool)

	handlers := &handler.Handlers{
		Auth:        handler.NewAuthHandler(authService, pool),
		Content:     handler.NewContentHandler(contentService),
		Lesson:      handler.NewLessonHandler(lessonService),
		Leaderboard: handler.NewLeaderboardHandler(leaderboardService),
		Admin:       handler.NewAdminHandler(adminService),
	}

	adminAuth := handler.AdminAuth{Username: cfg.AdminUsername, PasswordHash: cfg.AdminPasswordHash}
	if adminAuth.Username == "" || adminAuth.PasswordHash == "" {
		log.Println("admin panel disabled: set ADMIN_USERNAME and ADMIN_PASSWORD_HASH to enable /admin")
	}

	router := handler.NewRouter(handlers, authService, adminAuth)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("quranlingo api listening on :%s (%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
