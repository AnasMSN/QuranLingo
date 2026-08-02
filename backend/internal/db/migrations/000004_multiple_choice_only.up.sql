-- Restrict exercises to multiple-choice only, and add a draft/approved status
-- so admin-authored questions can be previewed before going live to the app.

-- Legacy 'translate' rows (from the original seeder) become multiple_choice;
-- `make backend-seed` backfills their missing options after this migration runs.
UPDATE exercises SET type = 'multiple_choice' WHERE type = 'translate';

ALTER TABLE exercises DROP CONSTRAINT exercises_type_check;
ALTER TABLE exercises ADD CONSTRAINT exercises_type_check CHECK (type = 'multiple_choice');

ALTER TABLE exercises ADD COLUMN status text NOT NULL DEFAULT 'approved' CHECK (status IN ('draft', 'approved'));

CREATE INDEX idx_exercises_status ON exercises(lesson_id, status);
