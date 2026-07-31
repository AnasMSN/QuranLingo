package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"quranlingo/backend/internal/models"
)

func CreateUser(ctx context.Context, q Querier, email, passwordHash, displayName string) (*models.User, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, display_name, total_xp, hearts, hearts_refill_at,
		          current_streak, longest_streak, last_activity_date, created_at, updated_at
	`, email, passwordHash, displayName)
	return scanUser(row)
}

func GetUserByEmail(ctx context.Context, q Querier, email string) (*models.User, error) {
	row := q.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, total_xp, hearts, hearts_refill_at,
		       current_streak, longest_streak, last_activity_date, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func GetUserByID(ctx context.Context, q Querier, id string) (*models.User, error) {
	row := q.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, total_xp, hearts, hearts_refill_at,
		       current_streak, longest_streak, last_activity_date, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

// UpdateUserAfterLesson persists the post-submission state of a user's XP,
// hearts, and streak. Callers must run this inside the same transaction as
// the lesson-completion and XP-transaction inserts.
func UpdateUserAfterLesson(ctx context.Context, q Querier, userID string, totalXP, hearts int, heartsRefillAt *time.Time, currentStreak, longestStreak int, lastActivityDate time.Time) error {
	_, err := q.Exec(ctx, `
		UPDATE users
		SET total_xp = $2, hearts = $3, hearts_refill_at = $4,
		    current_streak = $5, longest_streak = $6, last_activity_date = $7, updated_at = now()
		WHERE id = $1
	`, userID, totalXP, hearts, heartsRefillAt, currentStreak, longestStreak, lastActivityDate)
	return err
}

// ListUsers returns every user, newest first, for the admin panel.
func ListUsers(ctx context.Context, q Querier) ([]models.User, error) {
	rows, err := q.Query(ctx, `
		SELECT id, email, password_hash, display_name, total_xp, hearts, hearts_refill_at,
		       current_streak, longest_streak, last_activity_date, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

// RefillAllHearts resets every user below maxHearts (or with a pending
// lazy-refill timer) to full hearts. Returns how many rows were touched.
func RefillAllHearts(ctx context.Context, q Querier, maxHearts int) (int64, error) {
	tag, err := q.Exec(ctx, `
		UPDATE users
		SET hearts = $1, hearts_refill_at = NULL, updated_at = now()
		WHERE hearts < $1 OR hearts_refill_at IS NOT NULL
	`, maxHearts)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// RefillUserHearts resets a single user to full hearts.
func RefillUserHearts(ctx context.Context, q Querier, userID string, maxHearts int) error {
	_, err := q.Exec(ctx, `
		UPDATE users
		SET hearts = $2, hearts_refill_at = NULL, updated_at = now()
		WHERE id = $1
	`, userID, maxHearts)
	return err
}

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.TotalXP, &u.Hearts, &u.HeartsRefillAt,
		&u.CurrentStreak, &u.LongestStreak, &u.LastActivityDate, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
