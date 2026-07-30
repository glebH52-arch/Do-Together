DROP INDEX IF EXISTS project_members_one_creator_idx;
DROP INDEX IF EXISTS project_members_user_project_idx;

ALTER TABLE project_members
DROP CONSTRAINT IF EXISTS project_members_role_allowed_check;