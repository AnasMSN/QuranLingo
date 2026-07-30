package repository

import (
	"context"

	"quranlingo/backend/internal/models"
)

// WeeklyLeaderboard ranks users by XP earned in the last 7 days.
func WeeklyLeaderboard(ctx context.Context, q Querier, limit int) ([]models.LeaderboardEntry, error) {
	rows, err := q.Query(ctx, `
		SELECT u.id, u.display_name, COALESCE(SUM(x.amount), 0) AS weekly_xp
		FROM users u
		JOIN xp_transactions x ON x.user_id = u.id AND x.created_at >= now() - interval '7 days'
		GROUP BY u.id, u.display_name
		ORDER BY weekly_xp DESC, u.display_name ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.DisplayName, &e.WeeklyXP); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
