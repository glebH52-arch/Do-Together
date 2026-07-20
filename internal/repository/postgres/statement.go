package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	stmtCreateUser     = "create_user"
	stmtGetUserByID    = "get_user_by_id"
	stmtGetUserByEmail = "get_user_by_email"
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
