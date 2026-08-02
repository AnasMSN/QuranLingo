package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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

// --- Question management: draft -> preview -> approve ---
//
// Admin-authored questions (manual form or CSV import) are always created
// with status "draft" so an operator can preview them exactly as the app
// would render them, correct mistakes, then explicitly approve — at which
// point (and only then) they become visible to ContentService/LessonService.

var (
	ErrQuestionPromptRequired = errors.New("prompt is required")
	ErrQuestionOptionsInvalid = errors.New("provide 2-4 options with at least one marked correct")
	ErrQuestionNotDraft       = errors.New("question is not a draft")
)

// LessonOption is one entry in the admin "which lesson does this question
// belong to" picker.
type LessonOption struct {
	ID    string
	Label string
}

// QuestionOption is one multiple-choice option supplied by an admin, before
// it has a database ID.
type QuestionOption struct {
	Text    string
	Correct bool
}

// QuestionPreview is a draft (or approved) question plus its options, shaped
// for the admin preview page to render exactly like the mobile lesson screen.
type QuestionPreview struct {
	Exercise models.Exercise
	Options  []models.ExerciseOption
}

func (s *AdminService) ListLessonsForPicker(ctx context.Context) ([]LessonOption, error) {
	lessons, err := repository.ListLessonsForAdmin(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out := make([]LessonOption, len(lessons))
	for i, l := range lessons {
		out[i] = LessonOption{ID: l.LessonID, Label: fmt.Sprintf("%s — %s", l.SkillTitle, l.LessonTitle)}
	}
	return out, nil
}

func (s *AdminService) ListDraftQuestions(ctx context.Context, lessonID string) ([]QuestionPreview, error) {
	return s.listQuestions(ctx, lessonID, models.ExerciseStatusDraft)
}

// ListApprovedQuestions lists a lesson's live (approved) questions, for the
// "existing questions" management view — see and edit what's already in the app.
func (s *AdminService) ListApprovedQuestions(ctx context.Context, lessonID string) ([]QuestionPreview, error) {
	return s.listQuestions(ctx, lessonID, models.ExerciseStatusApproved)
}

func (s *AdminService) listQuestions(ctx context.Context, lessonID string, status models.ExerciseStatus) ([]QuestionPreview, error) {
	exercises, err := repository.ListExercisesByLessonAndStatus(ctx, s.db, lessonID, status)
	if err != nil {
		return nil, err
	}
	previews := make([]QuestionPreview, 0, len(exercises))
	for _, e := range exercises {
		opts, err := repository.ListOptionsByExercise(ctx, s.db, e.ID)
		if err != nil {
			return nil, err
		}
		previews = append(previews, QuestionPreview{Exercise: e, Options: opts})
	}
	return previews, nil
}

// validateQuestionOptions requires 2-4 non-empty options with at least one
// marked correct. Arabic words often have several valid English glosses (a
// general sense and one or more specific ones), so a question isn't limited
// to a single correct option — any option marked correct is graded as
// correct if the learner picks it (see gradeAnswer in lesson_service.go).
func validateQuestionOptions(options []QuestionOption) error {
	if len(options) < 2 || len(options) > 4 {
		return ErrQuestionOptionsInvalid
	}
	hasCorrect := false
	for _, o := range options {
		if strings.TrimSpace(o.Text) == "" {
			return ErrQuestionOptionsInvalid
		}
		if o.Correct {
			hasCorrect = true
		}
	}
	if !hasCorrect {
		return ErrQuestionOptionsInvalid
	}
	return nil
}

// correctAnswerText joins every acceptable answer for display in post-lesson
// results (ExerciseResult.CorrectAnswer) — e.g. "the Most Gracious / the Most Merciful".
func correctAnswerText(options []QuestionOption) string {
	var correct []string
	for _, o := range options {
		if o.Correct {
			correct = append(correct, o.Text)
		}
	}
	return strings.Join(correct, " / ")
}

// CreateDraftQuestion adds one new question to a lesson as a draft. audioURL
// is optional and stored as-is, unvalidated — a blank or unreachable URL is
// an expected "no clip recorded yet" state, not an error; playback (admin
// preview and the mobile app) just silently produces no sound in that case.
func (s *AdminService) CreateDraftQuestion(ctx context.Context, lessonID, prompt, arabicText, audioURL string, options []QuestionOption) (*models.Exercise, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, ErrQuestionPromptRequired
	}
	if err := validateQuestionOptions(options); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	position, err := repository.NextExercisePosition(ctx, tx, lessonID)
	if err != nil {
		return nil, err
	}

	exercise, err := repository.CreateExercise(ctx, tx, lessonID, prompt, strings.TrimSpace(arabicText), correctAnswerText(options), strings.TrimSpace(audioURL), position, models.ExerciseStatusDraft)
	if err != nil {
		return nil, err
	}
	for i, o := range options {
		if err := repository.CreateExerciseOption(ctx, tx, exercise.ID, strings.TrimSpace(o.Text), o.Correct, i); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return exercise, nil
}

// UpdateQuestion overwrites a question's prompt/Arabic text/options — used
// both to fix a mistake in a draft before approving it, and to correct an
// already-live question directly from the "existing questions" view. Edits
// to a live question take effect immediately, with no re-approval step.
func (s *AdminService) UpdateQuestion(ctx context.Context, exerciseID, prompt, arabicText, audioURL string, options []QuestionOption) error {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return ErrQuestionPromptRequired
	}
	if err := validateQuestionOptions(options); err != nil {
		return err
	}

	if _, err := repository.GetExerciseByID(ctx, s.db, exerciseID); err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	if err := repository.UpdateExercise(ctx, tx, exerciseID, prompt, strings.TrimSpace(arabicText), correctAnswerText(options), strings.TrimSpace(audioURL)); err != nil {
		return err
	}
	if err := repository.DeleteOptionsByExercise(ctx, tx, exerciseID); err != nil {
		return err
	}
	for i, o := range options {
		if err := repository.CreateExerciseOption(ctx, tx, exerciseID, strings.TrimSpace(o.Text), o.Correct, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ApproveQuestion publishes a draft, making it visible to the mobile app.
// Re-validates the option set server-side as a last line of defense — a
// draft can only ever be approved if it's actually gradable.
func (s *AdminService) ApproveQuestion(ctx context.Context, exerciseID string) error {
	options, err := repository.ListOptionsByExercise(ctx, s.db, exerciseID)
	if err != nil {
		return err
	}
	qOpts := make([]QuestionOption, len(options))
	for i, o := range options {
		qOpts[i] = QuestionOption{Text: o.OptionText, Correct: o.IsCorrect}
	}
	if err := validateQuestionOptions(qOpts); err != nil {
		return err
	}
	return repository.SetExerciseStatus(ctx, s.db, exerciseID, models.ExerciseStatusApproved)
}

// DeleteDraftQuestion rejects/removes a draft question. Never applies to
// already-approved (live) questions.
func (s *AdminService) DeleteDraftQuestion(ctx context.Context, exerciseID string) error {
	existing, err := repository.GetExerciseByID(ctx, s.db, exerciseID)
	if err != nil {
		return err
	}
	if existing.Status != models.ExerciseStatusDraft {
		return ErrQuestionNotDraft
	}
	return repository.DeleteExercise(ctx, s.db, exerciseID)
}

// ImportResult summarizes a CSV bulk-import: how many rows became draft
// questions, and a human-readable error per row that was skipped.
type ImportResult struct {
	Created int
	Errors  []string
}

// parseCorrectOptionNumbers parses a correct_option CSV cell into the
// 1-based option_N numbers it marks correct. Multiple acceptable answers
// (an Arabic word's general and more specific senses, say) are separated by
// ";" — e.g. "1;3" — since a plain "," would collide with CSV's own
// delimiter inside an unquoted cell.
func parseCorrectOptionNumbers(raw string) ([]int, error) {
	var nums []int
	for part := range strings.SplitSeq(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf(`correct_option must be number(s) separated by ";" (e.g. "1" or "1;3")`)
		}
		nums = append(nums, n)
	}
	if len(nums) == 0 {
		return nil, fmt.Errorf("correct_option is required")
	}
	return nums, nil
}

// ImportQuestionsCSV bulk-creates draft questions from a CSV file. Expected
// header columns: prompt, arabic_text (optional), option_1..option_4
// (2-4 required, Arabic with harakat is just UTF-8 text — no special
// handling needed), correct_option (1-based option_N number(s); separate
// multiple acceptable answers with ";", e.g. "1;3"), audio_url (optional —
// a blank or unreachable URL just means no clip yet, not an import error).
// Invalid rows are skipped and reported rather than aborting the whole
// import, since a bulk paste is likely to have a typo or two.
func (s *AdminService) ImportQuestionsCSV(ctx context.Context, lessonID string, r io.Reader) (*ImportResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // strip UTF-8 BOM (common from Excel exports)

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, name := range []string{"prompt", "option_1", "option_2", "correct_option"} {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("CSV is missing required column %q", name)
		}
	}

	get := func(row []string, name string) string {
		idx, ok := col[name]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	position, err := repository.NextExercisePosition(ctx, tx, lessonID)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{}
	rowNum := 1 // the header line
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNum, err))
			continue
		}
		if len(row) == 1 && strings.TrimSpace(row[0]) == "" {
			continue // blank line
		}

		prompt := get(row, "prompt")
		if prompt == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: prompt is required", rowNum))
			continue
		}

		type numberedOption struct {
			num  int
			text string
		}
		var numbered []numberedOption
		for n := 1; n <= 4; n++ {
			if text := get(row, fmt.Sprintf("option_%d", n)); text != "" {
				numbered = append(numbered, numberedOption{num: n, text: text})
			}
		}
		if len(numbered) < 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: provide at least 2 non-empty option_N columns", rowNum))
			continue
		}

		correctNums, err := parseCorrectOptionNumbers(get(row, "correct_option"))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNum, err))
			continue
		}
		correctSet := map[int]bool{}
		for _, n := range correctNums {
			correctSet[n] = true
		}

		options := make([]QuestionOption, len(numbered))
		foundCorrect := false
		for i, no := range numbered {
			correct := correctSet[no.num]
			foundCorrect = foundCorrect || correct
			options[i] = QuestionOption{Text: no.text, Correct: correct}
		}
		if !foundCorrect {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: correct_option doesn't match any non-empty option column", rowNum))
			continue
		}

		exercise, err := repository.CreateExercise(ctx, tx, lessonID, prompt, get(row, "arabic_text"), correctAnswerText(options), get(row, "audio_url"), position, models.ExerciseStatusDraft)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNum, err))
			continue
		}
		optionErr := false
		for i, o := range options {
			if err := repository.CreateExerciseOption(ctx, tx, exercise.ID, o.Text, o.Correct, i); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNum, err))
				optionErr = true
				break
			}
		}
		if optionErr {
			continue
		}
		position++
		result.Created++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
