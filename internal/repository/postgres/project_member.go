package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"

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

func (p *PostgresProjectMemberRepository) Add(ctx context.Context, actorID int, member *domain.ProjectMember) error {
	if member == nil {
		return ErrNilProjectMember
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := p.pool.Exec(ctx, stmtAddProjectMember, member.ProjectID, member.UserID, member.Role, member.JoinedAt, actorID)
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
	return nil
}
func (p *PostgresProjectMemberRepository) List(ctx context.Context, actorID, projectID int) ([]*domain.ProjectMember, error) {
	if actorID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if projectID <= 0 {
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

		err := rows.Scan(&projectMember.ProjectID, &projectMember.UserID, &projectMember.Username, &projectMember.Email, &projectMember.Role, &projectMember.JoinedAt)
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
func (p *PostgresProjectMemberRepository) UpdateRole(ctx context.Context, actorID, projectID, targetUserID int, role domain.ProjectMemberRole) error {
	if actorID <= 0 {
		return repository.ErrProjectNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if targetUserID <= 0 {
		return repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := p.pool.Exec(ctx, stmtUpdateRole, role, targetUserID, projectID, actorID)
	if err != nil {

		return fmt.Errorf("update project member: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrProjectNotFound
	}
	return nil
}
func (p *PostgresProjectMemberRepository) Remove(ctx context.Context, actorID, projectID, targetUserID int) error {
	if actorID <= 0 {
		return repository.ErrProjectNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if targetUserID <= 0 {
		return repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := p.pool.Exec(ctx, stmtRemoveProjectMember, targetUserID, projectID, actorID)
	if err != nil {

		return fmt.Errorf("remove project member: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrProjectNotFound
	}
	return nil
}
