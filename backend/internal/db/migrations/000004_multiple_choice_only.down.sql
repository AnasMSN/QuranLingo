DROP INDEX IF EXISTS idx_exercises_status;
ALTER TABLE exercises DROP COLUMN status;

ALTER TABLE exercises DROP CONSTRAINT exercises_type_check;
ALTER TABLE exercises ADD CONSTRAINT exercises_type_check CHECK (type IN ('multiple_choice', 'translate'));
