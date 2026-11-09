DROP INDEX IF EXISTS tasks_filter_idx;
DROP INDEX IF EXISTS tasks_description_trgm_idx;
DROP INDEX IF EXISTS tasks_title_trgm_idx;
DROP EXTENSION IF EXISTS pg_trgm;
