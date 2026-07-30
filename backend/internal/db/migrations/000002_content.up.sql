CREATE TABLE courses (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code           text NOT NULL UNIQUE,
    language_code  text NOT NULL DEFAULT 'ar',
    title          text NOT NULL,
    description    text,
    position       integer NOT NULL DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE skills (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id    uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    code         text NOT NULL,
    title        text NOT NULL,
    description  text,
    icon         text,
    position     integer NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (course_id, code),
    UNIQUE (course_id, position)
);

CREATE TABLE lessons (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id    uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    title       text NOT NULL,
    position    integer NOT NULL DEFAULT 0,
    xp_reward   integer NOT NULL DEFAULT 10,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (skill_id, position)
);

CREATE TABLE exercises (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id       uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    type            text NOT NULL CHECK (type IN ('multiple_choice', 'translate')),
    prompt          text NOT NULL,
    arabic_text     text,
    correct_answer  text NOT NULL,
    position        integer NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (lesson_id, position)
);

CREATE TABLE exercise_options (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_id  uuid NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    option_text  text NOT NULL,
    is_correct   boolean NOT NULL DEFAULT false,
    position     integer NOT NULL DEFAULT 0,
    UNIQUE (exercise_id, position)
);

CREATE INDEX idx_skills_course_id ON skills(course_id);
CREATE INDEX idx_lessons_skill_id ON lessons(skill_id);
CREATE INDEX idx_exercises_lesson_id ON exercises(lesson_id);
CREATE INDEX idx_exercise_options_exercise_id ON exercise_options(exercise_id);
