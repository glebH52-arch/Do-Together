package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.ProjectMemberRepository = (*PostgresProjectMemberRepository)(nil)

type PostgresProjectMemberRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProjectMemberRepository(pool *pgxpool.Pool) *PostgresProjectMemberRepository {
	return &PostgresProjectMemberRepository{
		pool: pool,
	}

}

var (
	ErrNilProjectMember = errors.New("project member is nil")
)

func (p *PostgresProjectMemberRepository) Add(ctx context.Context, actorID uuid.UUID, member *domain.ProjectMember) error {
	if member == nil {
		return ErrNilProjectMember
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin add project member: %w", err)
	}
	defer tx.Rollback(ctx)
	commandTag, err := tx.Exec(ctx, stmtAddProjectMember, member.ProjectID, member.UserID, member.Role, member.JoinedAt, actorID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return repository.ErrProjectMemberAlreadyExists
			case "23503":
				return repository.ErrUserNotFound
			}
		}
		return fmt.Errorf("insert project member: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrProjectNotFound
	}
	var targetName string
	if err := tx.QueryRow(ctx, `SELECT display_name FROM users WHERE id = $1`, member.UserID).Scan(&targetName); err != nil {
		return fmt.Errorf("get added member presentation: %w", err)
	}
	if err := insertProjectActivity(ctx, tx, member.ProjectID, actorID, member.UserID, "member.joined", "member", targetName, nil, map[string]any{
		"user_id": member.UserID, "display_name": targetName, "role": member.Role,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit add project member: %w", err)
	}
	return nil
}
func (p *PostgresProjectMemberRepository) List(ctx context.Context, actorID, projectID uuid.UUID) ([]*domain.ProjectMember, error) {
	if actorID == uuid.Nil() {
		return nil, repository.ErrProjectNotFound
	}
	if projectID == uuid.Nil() {
		return nil, repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, stmtGetListProjectMembers, projectID, actorID)

	if err != nil {
		return nil, fmt.Errorf("query list project member: %w", err)
	}

	defer rows.Close()

	projectMembers := make([]*domain.ProjectMember, 0)
	for rows.Next() {
		var projectMember domain.ProjectMember

		err := rows.Scan(&projectMember.ProjectID, &projectMember.UserID, &projectMember.Username, &projectMember.DisplayName, &projectMember.AvatarFilename, &projectMember.AvatarUpdatedAt, &projectMember.Email, &projectMember.Role, &projectMember.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		projectMembers = append(projectMembers, &projectMember)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project members: %w", err)
	}
	if len(projectMembers) == 0 {
		return nil, repository.ErrProjectNotFound
	}
	return projectMembers, nil
}
func (p *PostgresProjectMemberRepository) UpdateRole(ctx context.Context, actorID, projectID, targetUserID uuid.UUID, role domain.ProjectMemberRole) error {
	if actorID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if projectID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if targetUserID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin update project member: %w", err)
	}
	defer tx.Rollback(ctx)
	var oldRole domain.ProjectMemberRole
	var targetName string
	err = tx.QueryRow(ctx, `SELECT member.role, target.display_name
		FROM project_members AS member JOIN users AS target ON target.id = member.user_id
		WHERE member.project_id = $1 AND member.user_id = $2 FOR UPDATE OF member`, projectID, targetUserID).Scan(&oldRole, &targetName)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrProjectNotFound
	}
	if err != nil {
		return fmt.Errorf("lock project member: %w", err)
	}
	if oldRole == role {
		return nil
	}
	commandTag, err := tx.Exec(ctx, stmtUpdateRole, role, targetUserID, projectID, actorID)
	if err != nil {

		return fmt.Errorf("update project member: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrProjectNotFound
	}
	if err := insertProjectActivity(ctx, tx, projectID, actorID, targetUserID, "member.role_changed", "member", targetName,
		map[string]any{"role": oldRole, "user_id": targetUserID, "display_name": targetName},
		map[string]any{"role": role, "user_id": targetUserID, "display_name": targetName}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update project member: %w", err)
	}
	return nil
}
func (p *PostgresProjectMemberRepository) Remove(ctx context.Context, actorID, projectID, targetUserID uuid.UUID) error {
	if actorID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if projectID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if targetUserID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin remove project member: %w", err)
	}
	defer tx.Rollback(ctx)
	var oldRole domain.ProjectMemberRole
	var targetName string
	err = tx.QueryRow(ctx, `SELECT member.role, target.display_name
		FROM project_members AS member JOIN users AS target ON target.id = member.user_id
		WHERE member.project_id = $1 AND member.user_id = $2 FOR UPDATE OF member`, projectID, targetUserID).Scan(&oldRole, &targetName)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrProjectNotFound
	}
	if err != nil {
		return fmt.Errorf("lock project member for removal: %w", err)
	}
	commandTag, err := tx.Exec(ctx, stmtRemoveProjectMember, targetUserID, projectID, actorID)
	if err != nil {

		return fmt.Errorf("remove project member: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrProjectNotFound
	}
	if err := insertProjectActivity(ctx, tx, projectID, actorID, targetUserID, "member.removed", "member", targetName,
		map[string]any{"role": oldRole, "user_id": targetUserID, "display_name": targetName}, nil); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit remove project member: %w", err)
	}
	return nil
}
