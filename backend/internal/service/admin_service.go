package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
)

// AdminService backs the operator-only /admin panel: listing users and
// force-refilling hearts without waiting out the normal 4h lazy refill.
type AdminService struct {
	db *pgxpool.Pool
}

func NewAdminService(db *pgxpool.Pool) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) ListUsers(ctx context.Context) ([]models.User, error) {
	return repository.ListUsers(ctx, s.db)
}

// RefillAllHearts resets every user to full hearts. Returns how many users
// were actually touched (users already at full hearts are left alone).
func (s *AdminService) RefillAllHearts(ctx context.Context) (int64, error) {
	return repository.RefillAllHearts(ctx, s.db, MaxHearts)
}

func (s *AdminService) RefillUserHearts(ctx context.Context, userID string) error {
	return repository.RefillUserHearts(ctx, s.db, userID, MaxHearts)
}
