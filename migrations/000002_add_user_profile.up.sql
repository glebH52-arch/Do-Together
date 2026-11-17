ALTER TABLE users
    ADD COLUMN display_name TEXT,
    ADD COLUMN avatar_filename TEXT,
    ADD COLUMN avatar_updated_at TIMESTAMPTZ;

UPDATE users
SET display_name = username
WHERE display_name IS NULL;

ALTER TABLE users
    ALTER COLUMN display_name SET NOT NULL,
    ADD CONSTRAINT users_display_name_not_blank
        CHECK (BTRIM(display_name) <> ''),
    ADD CONSTRAINT users_avatar_filename_safe
        CHECK (
            avatar_filename IS NULL
            OR avatar_filename ~ '^[0-9a-f-]+\.(jpg|png)$'
        );
