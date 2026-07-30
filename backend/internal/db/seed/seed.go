// Package seed populates the "Arabic Basics" starter course. Every insert is
// an upsert keyed on a natural code/position, so running it repeatedly is safe.
package seed

import (
	"context"
	"fmt"

	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
)

func Run(ctx context.Context, q repository.Querier) error {
	course, err := repository.UpsertCourse(ctx, q, CourseCode, "ar", CourseTitle, CourseDescription, 0)
	if err != nil {
		return fmt.Errorf("upsert course: %w", err)
	}

	for skillPos, skill := range Skills {
		s, err := repository.UpsertSkill(ctx, q, course.ID, skill.Code, skill.Title, skill.Description, skill.Icon, skillPos)
		if err != nil {
			return fmt.Errorf("upsert skill %s: %w", skill.Code, err)
		}

		lesson, err := repository.UpsertLesson(ctx, q, s.ID, skill.Title, 0, 10)
		if err != nil {
			return fmt.Errorf("upsert lesson for skill %s: %w", skill.Code, err)
		}

		english := make([]string, len(skill.Words))
		for i, w := range skill.Words {
			english[i] = w.English
		}

		for i, w := range skill.Words {
			var (
				exType models.ExerciseType
				prompt string
			)
			if i%2 == 0 {
				exType = models.ExerciseMultipleChoice
				prompt = fmt.Sprintf("What does %q mean?", w.Arabic)
			} else {
				exType = models.ExerciseTranslate
				prompt = fmt.Sprintf("Type the English translation of %q", w.Arabic)
			}

			exercise, err := repository.UpsertExercise(ctx, q, lesson.ID, exType, prompt, w.Arabic, w.English, i)
			if err != nil {
				return fmt.Errorf("upsert exercise %d for skill %s: %w", i, skill.Code, err)
			}

			if exType == models.ExerciseMultipleChoice {
				options := optionsFor(i, w.English, english, 4)
				for optPos, opt := range options {
					if err := repository.UpsertExerciseOption(ctx, q, exercise.ID, opt, opt == w.English, optPos); err != nil {
						return fmt.Errorf("upsert option %d for exercise %d: %w", optPos, i, err)
					}
				}
			}
		}
	}

	return nil
}

// optionsFor builds a deterministic multiple-choice option set: the correct
// answer plus distractors drawn from other words in the same skill, in an
// order that rotates with i so the correct answer isn't always in the same slot.
func optionsFor(i int, correct string, pool []string, count int) []string {
	var distractors []string
	n := len(pool)
	for k := 1; k < n && len(distractors) < count-1; k++ {
		candidate := pool[(i+k)%n]
		if candidate != correct {
			distractors = append(distractors, candidate)
		}
	}

	options := append(distractors, correct)
	rotate := i % len(options)
	return append(options[rotate:], options[:rotate]...)
}
