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
