package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateUser            = "create_user"
	stmtGetUserByID           = "get_user_by_id"
	stmtGetUserByEmail        = "get_user_by_email"
	stmtCreateProject         = "create_project"
	stmtCreateProjectMember   = "create_project_member"
	stmtGetProjectByID        = "get_project_by_id"
	stmtGetListProjects       = "get_list_projects"
	stmtUpdateProject         = "update_project"
	stmtGetListProjectMembers = "get_list_project_members"
	stmtUpdateRole            = "update_role"
	stmtAddProjectMember      = "add_project_member"
	stmtRemoveProjectMember   = "remove_project_member"
	stmtCreateInvite          = "create_invite"
	stmtGetListInvites        = "get_list_invites"
	stmtAcceptInvite          = "accept_invite"
	stmtDeclineInvite         = "decline_invite"
	stmtGetInviteForUpdate    = "get_invite_for_update"
	stmtCreateTask            = "create_task"
	stmtGetListTask           = "get_list_task"
	stmtGetTaskByID           = "get_task_by_id"
	stmtUpdateTask            = "update_task"
	stmtRemoveTask            = "remove_task"
)

var statements = map[string]string{
	stmtCreateUser: `
		INSERT INTO users (username,email,password_hash,status,created_at)
		VALUES($1,$2,$3,$4,$5)
		RETURNING id
	`,
	stmtGetUserByID: `
		SELECT id,username,email,password_hash,status,created_at,updated_at FROM users
		WHERE id = $1`,
	stmtGetUserByEmail: `SELECT id,username,email,password_hash,status,created_at,updated_at FROM users
		WHERE email = $1`,
	stmtCreateProject: `
	INSERT INTO projects(created_by,title,goal,status,created_at)
	VALUES($1,$2,$3,$4,$5)
	RETURNING id
	`,
	stmtCreateProjectMember: `
	INSERT INTO project_members(project_id,user_id,role,joined_at)
	VALUES($1,$2,$3,$4)
	`,
	stmtGetProjectByID: `
		SELECT p.id,p.created_by,p.title,goal,p.status,p.created_at,p.updated_at FROM projects AS p
		JOIN project_members AS pm ON p.id = pm.project_id
		WHERE p.id = $1 AND pm.user_id = $2
	`,
	stmtGetListProjects: `
		SELECT p.id,p.created_by,p.title,goal,p.status,p.created_at,p.updated_at FROM projects AS p
		JOIN project_members AS pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND ($2::text IS NULL OR p.status = $2::text)
		ORDER BY p.id
		LIMIT  $3 OFFSET  $4
	`,
	stmtUpdateProject: `
	UPDATE projects AS p
	SET
    title = $1,
    goal = $2,
	status = $3,
    updated_at = $4
	FROM project_members AS pm
	WHERE p.id = $5
	AND pm.project_id = p.id
	AND pm.user_id = $6
	AND pm.role = 'creator'
	`,
	stmtGetListProjectMembers: `
SELECT
    pm.project_id,
    pm.user_id,
	 u.username,
    u.email,
    pm.role,
    pm.joined_at
FROM project_members pm
JOIN users AS u ON u.id = pm.user_id
WHERE pm.project_id = $1
AND EXISTS (
    SELECT 1
    FROM project_members AS access
    WHERE access.project_id = pm.project_id
      AND access.user_id = $2
)
ORDER BY pm.user_id
	`,
	stmtUpdateRole: `
	UPDATE project_members AS target
	SET
	role = $1
	WHERE target.user_id = $2
	AND target.project_id = $3
	AND target.role <> 'creator'
	AND $1 <> 'creator'
	AND EXISTS (
	SELECT 1
	FROM project_members AS actor
	WHERE actor.project_id = target.project_id
	AND actor.user_id = $4
	AND actor.role IN ('creator', 'admin')
	)
 `,
	stmtAddProjectMember: `
	INSERT INTO project_members(project_id,user_id,role,joined_at)
	SELECT $1, $2, $3, $4
	WHERE $3 <> 'creator'
	AND EXISTS (
	SELECT 1
    FROM project_members AS pm
	WHERE pm.project_id = $1
	AND pm.user_id = $5
	AND pm.role IN ('creator', 'admin')
)
	`,
	stmtRemoveProjectMember: `
	DELETE FROM project_members AS target
	WHERE target.user_id = $1
	AND target.project_id = $2
	AND target.role <> 'creator'
	AND EXISTS (
	SELECT 1
    FROM project_members AS actor
	WHERE actor.project_id = target.project_id
	AND actor.user_id = $3
	AND actor.role IN ('creator', 'admin')
)
	`,
	stmtCreateInvite: `
	INSERT INTO invites(project_id,inviter_id,invitee_id,role,status,expires_at,created_at)
	SELECT $1,$2,$3,$4,$5,$6,$7
	WHERE $4 IN ('admin', 'member')
	AND EXISTS (
	SELECT 1
    FROM project_members AS pm
	WHERE pm.project_id = $1
	AND pm.user_id = $2
	AND pm.role IN ('creator', 'admin')
)
	 AND NOT EXISTS (
      SELECT 1
      FROM project_members AS target
      WHERE target.project_id = $1
        AND target.user_id = $3
  )
RETURNING id
	`,
	stmtGetListInvites: `
	SELECT id,project_id,inviter_id,invitee_id,role,status,expires_at,created_at,updated_at FROM invites
	WHERE invitee_id = $1
	ORDER BY created_at , id
	`,
	stmtAcceptInvite: `
	UPDATE invites
SET status = $1,
    updated_at = $2
WHERE id = $3
  AND status = 'pending'
	`,
	stmtDeclineInvite: `
	UPDATE invites
SET status = 'declined',
    updated_at = $3
WHERE id = $1
  AND invitee_id = $2
  AND status = 'pending'
  AND expires_at > $3
	`,

	stmtGetInviteForUpdate: `
	SELECT id, project_id, inviter_id, invitee_id,
       role, status, expires_at, created_at, updated_at
FROM invites
WHERE id = $1
  AND invitee_id = $2
FOR UPDATE
	`,
	stmtCreateTask: `
	INSERT INTO tasks(project_id,title,description,created_by,status,created_at,due_at)
	SELECT $1,$2,$3,$4,$5,$6,$7
	WHERE EXISTS (
	SELECT 1
    FROM project_members AS pm
	WHERE pm.project_id = $1
	AND pm.user_id = $4
	)
	RETURNING id
	`,

	stmtUpdateTask: `
	UPDATE tasks AS target
	SET
    title = $1,
    description = $2,
    status = $3,
    updated_at = $4
	FROM project_members pm
	WHERE target.id = $5
	AND target.project_id = $6
	AND pm.project_id = target.project_id
	AND pm.user_id = $7;
	`,

	stmtGetListTask: `
	SELECT
    t.id, t.project_id, t.title, t.description,
    t.created_by, t.status, t.created_at,
    t.updated_at, t.due_at
FROM tasks AS t
WHERE t.project_id = $1
  AND EXISTS (
      SELECT 1
      FROM project_members AS pm
      WHERE pm.project_id = t.project_id
        AND pm.user_id = $2
  )
ORDER BY t.id
	`,

	stmtGetTaskByID: `
	SELECT
		t.id, t.project_id, t.title, t.description,
		t.created_by, t.status, t.created_at,
		t.updated_at, t.due_at
	FROM tasks AS t
	WHERE t.id = $1
	  AND t.project_id = $2
	  AND EXISTS (
		  SELECT 1
		  FROM project_members AS pm
		  WHERE pm.project_id = t.project_id
		    AND pm.user_id = $3
	  )
`,

	stmtRemoveTask: `
	DELETE FROM tasks AS target
	WHERE id = $1
	AND target.project_id = $2
	AND target.created_by = $3
	AND EXISTS (
	SELECT 1
	FROM project_members AS pm
	WHERE pm.project_id = target.project_id
	  AND pm.user_id = $3
)
	`,
}

func prepareStatements(ctx context.Context, conn *pgx.Conn) error {
	for name, sql := range statements {
		_, err := conn.Prepare(ctx, name, sql)
		if err != nil {
			return fmt.Errorf("prepare statement %s: %w", name, err)
		}
	}
	return nil
}
