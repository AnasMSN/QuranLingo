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

type ExerciseType string

const (
	ExerciseMultipleChoice ExerciseType = "multiple_choice"
	ExerciseTranslate      ExerciseType = "translate"
)

type Exercise struct {
	ID            string
	LessonID      string
	Type          ExerciseType
	Prompt        string
	ArabicText    string
	CorrectAnswer string
	Position      int
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
