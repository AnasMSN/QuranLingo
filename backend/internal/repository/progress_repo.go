package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"quranlingo/backend/internal/models"
)

// GetCompletedLessonIDs returns the subset of lessonIDs the user has already completed at least once.
func GetCompletedLessonIDs(ctx context.Context, q Querier, userID string, lessonIDs []string) (map[string]bool, error) {
	if len(lessonIDs) == 0 {
		return map[string]bool{}, nil
	}
	rows, err := q.Query(ctx, `
		SELECT DISTINCT lesson_id FROM user_lesson_completions
		WHERE user_id = $1 AND lesson_id = ANY($2)
	`, userID, lessonIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	completed := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		completed[id] = true
	}
	return completed, rows.Err()
}

// GetCompletionByIdempotencyKey looks up a prior submission so retried requests
// (flaky mobile connections) don't double-award XP.
func GetCompletionByIdempotencyKey(ctx context.Context, q Querier, userID, idempotencyKey string) (*models.LessonCompletion, error) {
	row := q.QueryRow(ctx, `
		SELECT id, user_id, lesson_id, score, xp_earned, idempotency_key, completed_at
		FROM user_lesson_completions WHERE user_id = $1 AND idempotency_key = $2
	`, userID, idempotencyKey)

	var c models.LessonCompletion
	err := row.Scan(&c.ID, &c.UserID, &c.LessonID, &c.Score, &c.XPEarned, &c.IdempotencyKey, &c.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func CreateLessonCompletion(ctx context.Context, q Querier, userID, lessonID string, score, xpEarned int, idempotencyKey string) (*models.LessonCompletion, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO user_lesson_completions (user_id, lesson_id, score, xp_earned, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, lesson_id, score, xp_earned, idempotency_key, completed_at
	`, userID, lessonID, score, xpEarned, idempotencyKey)

	var c models.LessonCompletion
	if err := row.Scan(&c.ID, &c.UserID, &c.LessonID, &c.Score, &c.XPEarned, &c.IdempotencyKey, &c.CompletedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateXPTransaction(ctx context.Context, q Querier, userID string, amount int, reason string, lessonCompletionID string) error {
	_, err := q.Exec(ctx, `
		INSERT INTO xp_transactions (user_id, amount, reason, lesson_completion_id)
		VALUES ($1, $2, $3, $4)
	`, userID, amount, reason, lessonCompletionID)
	return err
}
