package service

import (
	"context"
	"errors"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
)

var ErrLessonLocked = errors.New("lesson is locked")

const (
	StatusLocked    = "locked"
	StatusUnlocked  = "unlocked"
	StatusCompleted = "completed"
)

type ContentService struct {
	db *pgxpool.Pool
}

func NewContentService(db *pgxpool.Pool) *ContentService {
	return &ContentService{db: db}
}

type LessonNode struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Position int    `json:"position"`
	XPReward int    `json:"xp_reward"`
	Status   string `json:"status"`
}

type SkillNode struct {
	ID          string       `json:"id"`
	Code        string       `json:"code"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Position    int          `json:"position"`
	Status      string       `json:"status"`
	Lessons     []LessonNode `json:"lessons"`
}

type CourseTree struct {
	ID          string      `json:"id"`
	Code        string      `json:"code"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Skills      []SkillNode `json:"skills"`
}

// GetCourseTree builds the full course -> skills -> lessons tree with
// per-lesson/per-skill unlock status computed for the given user.
func (s *ContentService) GetCourseTree(ctx context.Context, userID, courseCode string) (*CourseTree, error) {
	course, err := repository.GetCourseByCode(ctx, s.db, courseCode)
	if err != nil {
		return nil, err
	}

	skills, err := repository.ListSkillsByCourse(ctx, s.db, course.ID)
	if err != nil {
		return nil, err
	}

	lessonsBySkill, completed, err := s.loadLessonsAndProgress(ctx, userID, skills)
	if err != nil {
		return nil, err
	}

	tree := &CourseTree{ID: course.ID, Code: course.Code, Title: course.Title, Description: course.Description}

	previousSkillCompleted := true
	for _, skill := range skills {
		skillLessons := lessonsBySkill[skill.ID]
		skillUnlocked := previousSkillCompleted

		lessonNodes := make([]LessonNode, 0, len(skillLessons))
		previousLessonCompleted := true
		allCompleted := len(skillLessons) > 0
		for _, lesson := range skillLessons {
			isCompleted := completed[lesson.ID]
			isUnlocked := skillUnlocked && previousLessonCompleted

			status := StatusLocked
			switch {
			case isCompleted:
				status = StatusCompleted
			case isUnlocked:
				status = StatusUnlocked
			}

			lessonNodes = append(lessonNodes, LessonNode{
				ID: lesson.ID, Title: lesson.Title, Position: lesson.Position,
				XPReward: lesson.XPReward, Status: status,
			})

			previousLessonCompleted = isCompleted
			if !isCompleted {
				allCompleted = false
			}
		}

		skillStatus := StatusLocked
		switch {
		case allCompleted:
			skillStatus = StatusCompleted
		case skillUnlocked:
			skillStatus = StatusUnlocked
		}

		tree.Skills = append(tree.Skills, SkillNode{
			ID: skill.ID, Code: skill.Code, Title: skill.Title, Description: skill.Description,
			Icon: skill.Icon, Position: skill.Position, Status: skillStatus, Lessons: lessonNodes,
		})

		previousSkillCompleted = allCompleted
	}

	return tree, nil
}

type ExerciseOptionDTO struct {
	ID         string `json:"id"`
	OptionText string `json:"option_text"`
}

type ExerciseDTO struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Prompt     string              `json:"prompt"`
	ArabicText string              `json:"arabic_text,omitempty"`
	AudioURL   string              `json:"audio_url,omitempty"`
	Options    []ExerciseOptionDTO `json:"options,omitempty"`
}

type LessonDetail struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	XPReward  int           `json:"xp_reward"`
	Exercises []ExerciseDTO `json:"exercises"`
}

// GetLessonDetail returns a lesson's exercises for the user to attempt.
// Correct answers and which multiple-choice option is correct are withheld —
// grading only ever happens server-side in LessonService.Submit.
func (s *ContentService) GetLessonDetail(ctx context.Context, userID, lessonID string) (*LessonDetail, error) {
	lesson, err := repository.GetLessonByID(ctx, s.db, lessonID)
	if err != nil {
		return nil, err
	}

	if err := s.assertLessonUnlocked(ctx, userID, lesson); err != nil {
		return nil, err
	}

	exercises, err := repository.ListExercisesByLesson(ctx, s.db, lesson.ID)
	if err != nil {
		return nil, err
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

	detail := &LessonDetail{ID: lesson.ID, Title: lesson.Title, XPReward: lesson.XPReward}
	for _, e := range exercises {
		dto := ExerciseDTO{ID: e.ID, Type: string(e.Type), Prompt: e.Prompt, ArabicText: e.ArabicText, AudioURL: e.AudioURL}
		for _, o := range optionsByExercise[e.ID] {
			dto.Options = append(dto.Options, ExerciseOptionDTO{ID: o.ID, OptionText: o.OptionText})
		}
		detail.Exercises = append(detail.Exercises, dto)
	}
	return detail, nil
}

func (s *ContentService) assertLessonUnlocked(ctx context.Context, userID string, lesson *models.Lesson) error {
	skill, err := repository.GetSkillByID(ctx, s.db, lesson.SkillID)
	if err != nil {
		return err
	}
	skills, err := repository.ListSkillsByCourse(ctx, s.db, skill.CourseID)
	if err != nil {
		return err
	}
	lessonsBySkill, completed, err := s.loadLessonsAndProgress(ctx, userID, skills)
	if err != nil {
		return err
	}

	previousSkillCompleted := true
	for _, sk := range skills {
		skillLessons := lessonsBySkill[sk.ID]
		skillUnlocked := previousSkillCompleted
		previousLessonCompleted := true
		allCompleted := len(skillLessons) > 0

		for _, l := range skillLessons {
			isCompleted := completed[l.ID]
			isUnlocked := skillUnlocked && previousLessonCompleted

			if l.ID == lesson.ID {
				if !isUnlocked && !isCompleted {
					return ErrLessonLocked
				}
				return nil
			}

			previousLessonCompleted = isCompleted
			if !isCompleted {
				allCompleted = false
			}
		}
		previousSkillCompleted = allCompleted
	}
	return ErrLessonLocked
}

func (s *ContentService) loadLessonsAndProgress(ctx context.Context, userID string, skills []models.Skill) (map[string][]models.Lesson, map[string]bool, error) {
	skillIDs := make([]string, len(skills))
	for i, sk := range skills {
		skillIDs[i] = sk.ID
	}

	lessons, err := repository.ListLessonsBySkillIDs(ctx, s.db, skillIDs)
	if err != nil {
		return nil, nil, err
	}

	lessonsBySkill := map[string][]models.Lesson{}
	lessonIDs := make([]string, len(lessons))
	for i, l := range lessons {
		lessonsBySkill[l.SkillID] = append(lessonsBySkill[l.SkillID], l)
		lessonIDs[i] = l.ID
	}
	for skillID := range lessonsBySkill {
		sort.Slice(lessonsBySkill[skillID], func(i, j int) bool {
			return lessonsBySkill[skillID][i].Position < lessonsBySkill[skillID][j].Position
		})
	}

	completed, err := repository.GetCompletedLessonIDs(ctx, s.db, userID, lessonIDs)
	if err != nil {
		return nil, nil, err
	}

	return lessonsBySkill, completed, nil
}
