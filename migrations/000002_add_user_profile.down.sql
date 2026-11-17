ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_avatar_filename_safe,
    DROP CONSTRAINT IF EXISTS users_display_name_not_blank,
    DROP COLUMN IF EXISTS avatar_updated_at,
    DROP COLUMN IF EXISTS avatar_filename,
    DROP COLUMN IF EXISTS display_name;
