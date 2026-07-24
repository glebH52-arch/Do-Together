package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateUser          = "create_user"
	stmtGetUserByID         = "get_user_by_id"
	stmtGetUserByEmail      = "get_user_by_email"
	stmtCreateProject       = "create_project"
	stmtCreateProjectMember = "create_project_member"
	stmtGetProjectByID      = "get_project_by_id"
	stmtGetListProjects     = "get_list_projects"
	stmtUpdateProject       = "update_project"
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
		WHERE pm.user_id = $1
		ORDER BY p.id
	`,

	stmtUpdateProject: `
	UPDATE projects AS p
	SET
    title = $1,
    goal = $2,
    updated_at = $3
	FROM project_members AS pm
	WHERE p.id = $4
	AND pm.project_id = p.id
	AND pm.user_id = $5
	AND pm.role = 'creator'
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
