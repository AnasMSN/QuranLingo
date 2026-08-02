package models

import "time"

type User struct {
	ID               string
	Email            string
	PasswordHash     string
	DisplayName      string
	TotalXP          int
	Hearts           int
	HeartsRefillAt   *time.Time
	CurrentStreak    int
	LongestStreak    int
	LastActivityDate *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Course struct {
	ID           string
	Code         string
	LanguageCode string
	Title        string
	Description  string
	Position     int
}

type Skill struct {
	ID          string
	CourseID    string
	Code        string
	Title       string
	Description string
	Icon        string
	Position    int
}

type Lesson struct {
	ID       string
	SkillID  string
	Title    string
	Position int
	XPReward int
}

// ExerciseType is always ExerciseMultipleChoice — the app only ever supports
// multiple-choice questions. The column/type still exist so a different
// exercise type could be reintroduced later without a schema rewrite.
type ExerciseType string

const (
	ExerciseMultipleChoice ExerciseType = "multiple_choice"
)

// ExerciseStatus gates whether a question is visible to the mobile app.
// Admin-authored questions (manual entry or CSV import) start as draft so an
// operator can preview and edit them before they go live.
type ExerciseStatus string

const (
	ExerciseStatusDraft    ExerciseStatus = "draft"
	ExerciseStatusApproved ExerciseStatus = "approved"
)

type Exercise struct {
	ID            string
	LessonID      string
	Type          ExerciseType
	Prompt        string
	ArabicText    string
	CorrectAnswer string
	Position      int
	Status        ExerciseStatus
	AudioURL      string
}

type ExerciseOption struct {
	ID         string
	ExerciseID string
	OptionText string
	IsCorrect  bool
	Position   int
}

type LessonCompletion struct {
	ID              string
	UserID          string
	LessonID        string
	Score           int
	XPEarned        int
	IdempotencyKey  string
	CompletedAt     time.Time
}

type LeaderboardEntry struct {
	UserID      string
	DisplayName string
	WeeklyXP    int
}

type XPTransaction struct {
	ID         string
	UserID     string
	Amount     int
	Reason     string
	CreatedAt  time.Time
}
