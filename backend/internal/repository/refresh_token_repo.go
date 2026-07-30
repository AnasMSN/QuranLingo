package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"quranlingo/backend/internal/models"
)

func CreateRefreshToken(ctx context.Context, q Querier, userID, tokenHash string, expiresAt time.Time) error {
	_, err := q.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func GetRefreshTokenByHash(ctx context.Context, q Querier, tokenHash string) (*models.RefreshToken, error) {
	row := q.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`, tokenHash)

	var rt models.RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func RevokeRefreshToken(ctx context.Context, q Querier, tokenHash string) error {
	_, err := q.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}
