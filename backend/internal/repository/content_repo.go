package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"quranlingo/backend/internal/models"
)

// --- Reads ---

func GetCourseByCode(ctx context.Context, q Querier, code string) (*models.Course, error) {
	row := q.QueryRow(ctx, `
		SELECT id, code, language_code, title, description, position
		FROM courses WHERE code = $1
	`, code)

	var c models.Course
	if err := row.Scan(&c.ID, &c.Code, &c.LanguageCode, &c.Title, &c.Description, &c.Position); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func GetSkillByID(ctx context.Context, q Querier, skillID string) (*models.Skill, error) {
	row := q.QueryRow(ctx, `
		SELECT id, course_id, code, title, description, icon, position
		FROM skills WHERE id = $1
	`, skillID)

	var s models.Skill
	if err := row.Scan(&s.ID, &s.CourseID, &s.Code, &s.Title, &s.Description, &s.Icon, &s.Position); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func ListSkillsByCourse(ctx context.Context, q Querier, courseID string) ([]models.Skill, error) {
	rows, err := q.Query(ctx, `
		SELECT id, course_id, code, title, description, icon, position
		FROM skills WHERE course_id = $1 ORDER BY position ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		if err := rows.Scan(&s.ID, &s.CourseID, &s.Code, &s.Title, &s.Description, &s.Icon, &s.Position); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}

func ListLessonsBySkillIDs(ctx context.Context, q Querier, skillIDs []string) ([]models.Lesson, error) {
	if len(skillIDs) == 0 {
		return nil, nil
	}
	rows, err := q.Query(ctx, `
		SELECT id, skill_id, title, position, xp_reward
		FROM lessons WHERE skill_id = ANY($1) ORDER BY skill_id, position ASC
	`, skillIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []models.Lesson
	for rows.Next() {
		var l models.Lesson
		if err := rows.Scan(&l.ID, &l.SkillID, &l.Title, &l.Position, &l.XPReward); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func GetLessonByID(ctx context.Context, q Querier, lessonID string) (*models.Lesson, error) {
	row := q.QueryRow(ctx, `
		SELECT id, skill_id, title, position, xp_reward
		FROM lessons WHERE id = $1
	`, lessonID)

	var l models.Lesson
	if err := row.Scan(&l.ID, &l.SkillID, &l.Title, &l.Position, &l.XPReward); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &l, nil
}

func ListExercisesByLesson(ctx context.Context, q Querier, lessonID string) ([]models.Exercise, error) {
	rows, err := q.Query(ctx, `
		SELECT id, lesson_id, type, prompt, arabic_text, correct_answer, position
		FROM exercises WHERE lesson_id = $1 ORDER BY position ASC
	`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var e models.Exercise
		var exType string
		if err := rows.Scan(&e.ID, &e.LessonID, &exType, &e.Prompt, &e.ArabicText, &e.CorrectAnswer, &e.Position); err != nil {
			return nil, err
		}
		e.Type = models.ExerciseType(exType)
		exercises = append(exercises, e)
	}
	return exercises, rows.Err()
}

func ListOptionsByExerciseIDs(ctx context.Context, q Querier, exerciseIDs []string) ([]models.ExerciseOption, error) {
	if len(exerciseIDs) == 0 {
		return nil, nil
	}
	rows, err := q.Query(ctx, `
		SELECT id, exercise_id, option_text, is_correct, position
		FROM exercise_options WHERE exercise_id = ANY($1) ORDER BY exercise_id, position ASC
	`, exerciseIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []models.ExerciseOption
	for rows.Next() {
		var o models.ExerciseOption
		if err := rows.Scan(&o.ID, &o.ExerciseID, &o.OptionText, &o.IsCorrect, &o.Position); err != nil {
			return nil, err
		}
		options = append(options, o)
	}
	return options, rows.Err()
}

// --- Upserts (used by the seed command; safe to re-run) ---

func UpsertCourse(ctx context.Context, q Querier, code, languageCode, title, description string, position int) (*models.Course, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO courses (code, language_code, title, description, position)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (code) DO UPDATE
		SET language_code = EXCLUDED.language_code, title = EXCLUDED.title,
		    description = EXCLUDED.description, position = EXCLUDED.position
		RETURNING id, code, language_code, title, description, position
	`, code, languageCode, title, description, position)

	var c models.Course
	if err := row.Scan(&c.ID, &c.Code, &c.LanguageCode, &c.Title, &c.Description, &c.Position); err != nil {
		return nil, err
	}
	return &c, nil
}

func UpsertSkill(ctx context.Context, q Querier, courseID, code, title, description, icon string, position int) (*models.Skill, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO skills (course_id, code, title, description, icon, position)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (course_id, code) DO UPDATE
		SET title = EXCLUDED.title, description = EXCLUDED.description,
		    icon = EXCLUDED.icon, position = EXCLUDED.position
		RETURNING id, course_id, code, title, description, icon, position
	`, courseID, code, title, description, icon, position)

	var s models.Skill
	if err := row.Scan(&s.ID, &s.CourseID, &s.Code, &s.Title, &s.Description, &s.Icon, &s.Position); err != nil {
		return nil, err
	}
	return &s, nil
}

func UpsertLesson(ctx context.Context, q Querier, skillID, title string, position, xpReward int) (*models.Lesson, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO lessons (skill_id, title, position, xp_reward)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (skill_id, position) DO UPDATE
		SET title = EXCLUDED.title, xp_reward = EXCLUDED.xp_reward
		RETURNING id, skill_id, title, position, xp_reward
	`, skillID, title, position, xpReward)

	var l models.Lesson
	if err := row.Scan(&l.ID, &l.SkillID, &l.Title, &l.Position, &l.XPReward); err != nil {
		return nil, err
	}
	return &l, nil
}

func UpsertExercise(ctx context.Context, q Querier, lessonID string, exType models.ExerciseType, prompt, arabicText, correctAnswer string, position int) (*models.Exercise, error) {
	row := q.QueryRow(ctx, `
		INSERT INTO exercises (lesson_id, type, prompt, arabic_text, correct_answer, position)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (lesson_id, position) DO UPDATE
		SET type = EXCLUDED.type, prompt = EXCLUDED.prompt,
		    arabic_text = EXCLUDED.arabic_text, correct_answer = EXCLUDED.correct_answer
		RETURNING id, lesson_id, type, prompt, arabic_text, correct_answer, position
	`, lessonID, string(exType), prompt, arabicText, correctAnswer, position)

	var e models.Exercise
	var scannedType string
	if err := row.Scan(&e.ID, &e.LessonID, &scannedType, &e.Prompt, &e.ArabicText, &e.CorrectAnswer, &e.Position); err != nil {
		return nil, err
	}
	e.Type = models.ExerciseType(scannedType)
	return &e, nil
}

func UpsertExerciseOption(ctx context.Context, q Querier, exerciseID, optionText string, isCorrect bool, position int) error {
	_, err := q.Exec(ctx, `
		INSERT INTO exercise_options (exercise_id, option_text, is_correct, position)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (exercise_id, position) DO UPDATE
		SET option_text = EXCLUDED.option_text, is_correct = EXCLUDED.is_correct
	`, exerciseID, optionText, isCorrect, position)
	return err
}
