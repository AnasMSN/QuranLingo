-- Per-question pronunciation clip. Nothing plays it yet from the admin side
-- unless an operator sets it (see /admin/questions); the mobile app treats a
-- blank/unreachable URL as "no audio available" rather than an error.
ALTER TABLE exercises ADD COLUMN audio_url text NOT NULL DEFAULT '';
