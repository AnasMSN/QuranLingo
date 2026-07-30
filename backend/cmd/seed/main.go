// Command seed populates the database with the "125 Words of the Qur'an"
// course. Safe to run multiple times — every insert is an upsert.
package main

import (
	"context"
	"log"
	"time"

	"quranlingo/backend/internal/config"
	"quranlingo/backend/internal/db"
	"quranlingo/backend/internal/db/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := seed.Run(ctx, pool); err != nil {
		log.Fatalf("seed: %v", err)
	}

	log.Printf("seed complete: course '%s' with %d skills is up to date", seed.CourseCode, len(seed.Skills))
}
