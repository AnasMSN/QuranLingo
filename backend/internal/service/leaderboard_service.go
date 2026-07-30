package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
)

type LeaderboardService struct {
	db *pgxpool.Pool
}

func NewLeaderboardService(db *pgxpool.Pool) *LeaderboardService {
	return &LeaderboardService{db: db}
}

func (s *LeaderboardService) Weekly(ctx context.Context, limit int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return repository.WeeklyLeaderboard(ctx, s.db, limit)
}
