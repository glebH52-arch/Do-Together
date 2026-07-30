CREATE UNIQUE INDEX project_members_one_creator_idx
ON project_members (project_id)
WHERE role = 'creator';

ALTER TABLE project_members
ADD CONSTRAINT project_members_role_allowed_check
CHECK (role IN ('creator', 'admin', 'member'));


CREATE INDEX project_members_user_project_idx
ON project_members (user_id, project_id);