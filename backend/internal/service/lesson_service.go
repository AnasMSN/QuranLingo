package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
)

const (
	// MaxHearts is exported so AdminService can refill users to full without
	// duplicating the value.
	MaxHearts          = 5
	heartsRefillPeriod = 4 * time.Hour
)

var (
	ErrNoHeartsRemaining  = errors.New("no hearts remaining")
	ErrIdempotencyKeyReq  = errors.New("idempotency_key is required")
	ErrLessonHasNoContent = errors.New("lesson has no exercises")
)

type LessonService struct {
	db *pgxpool.Pool
}

func NewLessonService(db *pgxpool.Pool) *LessonService {
	return &LessonService{db: db}
}

type AnswerInput struct {
	ExerciseID string `json:"exercise_id"`
	OptionID   string `json:"option_id,omitempty"`
}

type ExerciseResult struct {
	ExerciseID    string `json:"exercise_id"`
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correct_answer"`
}

type SubmitResult struct {
	AlreadySubmitted bool             `json:"already_submitted"`
	Score            int              `json:"score"`
	XPEarned         int              `json:"xp_earned"`
	TotalXP          int              `json:"total_xp"`
	HeartsRemaining  int              `json:"hearts_remaining"`
	CurrentStreak    int              `json:"current_streak"`
	LongestStreak    int              `json:"longest_streak"`
	Results          []ExerciseResult `json:"results,omitempty"`
}

// Submit grades a lesson attempt entirely server-side: the client only ever
// supplies raw answers, never a score. XP, hearts, and streak are all
// recomputed here and persisted atomically in one transaction.
func (s *LessonService) Submit(ctx context.Context, userID, lessonID, idempotencyKey string, answers []AnswerInput) (*SubmitResult, error) {
	if idempotencyKey == "" {
		return nil, ErrIdempotencyKeyReq
	}

	if existing, err := repository.GetCompletionByIdempotencyKey(ctx, s.db, userID, idempotencyKey); err == nil {
		user, err := repository.GetUserByID(ctx, s.db, userID)
		if err != nil {
			return nil, err
		}
		return &SubmitResult{
			AlreadySubmitted: true,
			Score:            existing.Score,
			XPEarned:         existing.XPEarned,
			TotalXP:          user.TotalXP,
			HeartsRemaining:  user.Hearts,
			CurrentStreak:    user.CurrentStreak,
			LongestStreak:    user.LongestStreak,
		}, nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	user, err := repository.GetUserByID(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	lesson, err := repository.GetLessonByID(ctx, s.db, lessonID)
	if err != nil {
		return nil, err
	}

	exercises, err := repository.ListExercisesByLesson(ctx, s.db, lessonID)
	if err != nil {
		return nil, err
	}
	if len(exercises) == 0 {
		return nil, ErrLessonHasNoContent
	}

	exerciseIDs := make([]string, len(exercises))
	for i, e := range exercises {
		exerciseIDs[i] = e.ID
	}
	options, err := repository.ListOptionsByExerciseIDs(ctx, s.db, exerciseIDs)
	if err != nil {
		return nil, err
	}
	optionsByExercise := map[string][]models.ExerciseOption{}
	for _, o := range options {
		optionsByExercise[o.ExerciseID] = append(optionsByExercise[o.ExerciseID], o)
	}

	answerByExercise := map[string]AnswerInput{}
	for _, a := range answers {
		answerByExercise[a.ExerciseID] = a
	}

	now := time.Now().UTC()

	hearts := user.Hearts
	heartsRefillAt := user.HeartsRefillAt
	if hearts <= 0 {
		if heartsRefillAt != nil && now.After(*heartsRefillAt) {
			hearts = MaxHearts
			heartsRefillAt = nil
		} else {
			return nil, ErrNoHeartsRemaining
		}
	}

	results := make([]ExerciseResult, 0, len(exercises))
	correctCount := 0
	for _, e := range exercises {
		correct := gradeAnswer(answerByExercise[e.ID], optionsByExercise[e.ID])
		if correct {
			correctCount++
		}
		results = append(results, ExerciseResult{ExerciseID: e.ID, Correct: correct, CorrectAnswer: e.CorrectAnswer})
	}

	score := correctCount * 100 / len(exercises)
	xpEarned := lesson.XPReward
	wrongCount := len(exercises) - correctCount

	hearts -= wrongCount
	if hearts < 0 {
		hearts = 0
	}
	if hearts == 0 && wrongCount > 0 {
		refillAt := now.Add(heartsRefillPeriod)
		heartsRefillAt = &refillAt
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	currentStreak, longestStreak := nextStreak(user.LastActivityDate, today, user.CurrentStreak, user.LongestStreak)

	newTotalXP := user.TotalXP + xpEarned

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	completion, err := repository.CreateLessonCompletion(ctx, tx, userID, lessonID, score, xpEarned, idempotencyKey)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Lost a race against a concurrent identical submission; the other
			// request's result is authoritative.
			_ = tx.Rollback(ctx)
			existing, ferr := repository.GetCompletionByIdempotencyKey(ctx, s.db, userID, idempotencyKey)
			if ferr != nil {
				return nil, ferr
			}
			freshUser, ferr := repository.GetUserByID(ctx, s.db, userID)
			if ferr != nil {
				return nil, ferr
			}
			return &SubmitResult{
				AlreadySubmitted: true,
				Score:            existing.Score,
				XPEarned:         existing.XPEarned,
				TotalXP:          freshUser.TotalXP,
				HeartsRemaining:  freshUser.Hearts,
				CurrentStreak:    freshUser.CurrentStreak,
				LongestStreak:    freshUser.LongestStreak,
			}, nil
		}
		return nil, err
	}

	if err := repository.CreateXPTransaction(ctx, tx, userID, xpEarned, "lesson_complete", completion.ID); err != nil {
		return nil, err
	}

	if err := repository.UpdateUserAfterLesson(ctx, tx, userID, newTotalXP, hearts, heartsRefillAt, currentStreak, longestStreak, today); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &SubmitResult{
		Score:           score,
		XPEarned:        xpEarned,
		TotalXP:         newTotalXP,
		HeartsRemaining: hearts,
		CurrentStreak:   currentStreak,
		LongestStreak:   longestStreak,
		Results:         results,
	}, nil
}

// gradeAnswer grades a multiple-choice answer — the only exercise type the
// app supports. An option is correct only if it matches one of the
// exercise's own options and that option is flagged is_correct in the DB.
func gradeAnswer(ans AnswerInput, options []models.ExerciseOption) bool {
	if ans.OptionID == "" {
		return false
	}
	for _, o := range options {
		if o.ID == ans.OptionID {
			return o.IsCorrect
		}
	}
	return false
}

func nextStreak(lastActivityDate *time.Time, today time.Time, currentStreak, longestStreak int) (int, int) {
	switch {
	case lastActivityDate == nil:
		currentStreak = 1
	case sameCalendarDay(*lastActivityDate, today):
		// already practiced today; streak unchanged
	case sameCalendarDay(*lastActivityDate, today.AddDate(0, 0, -1)):
		currentStreak++
	default:
		currentStreak = 1
	}
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}
	return currentStreak, longestStreak
}

func sameCalendarDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
