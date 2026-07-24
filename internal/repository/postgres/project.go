package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.ProjectRepository = (*PostgresProjectRepository)(nil)

type PostgresProjectRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProjectRepository(pool *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{
		pool: pool,
	}

}

func (p *PostgresProjectRepository) Create(ctx context.Context, userID int, project *domain.Project) error {
	if project == nil {
		return repository.ErrNilProject
	}
	if project.ID != 0 {
		return repository.ErrProjectAlreadySaved
	}
	if userID <= 0 {
		return repository.ErrUserNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin create project transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var id int
	err = tx.QueryRow(ctx, stmtCreateProject, userID, project.Title, project.Goal, project.Status, project.CreatedAt).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	_, err = tx.Exec(ctx, stmtCreateProjectMember, id, userID, "creator", time.Now())
	if err != nil {
		return fmt.Errorf("insert project member: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %ws", err)
	}
	project.ID = id
	project.CreatedBy = userID
	return nil
}
func (p *PostgresProjectRepository) GetByID(ctx context.Context, userID, id int) (*domain.Project, error) {
	if userID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	project := domain.Project{}

	err := p.pool.QueryRow(ctx, stmtGetProjectByID, id, userID).Scan(&project.ID, &project.CreatedBy, &project.Title, &project.Goal, &project.Status, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProjectNotFound
		}
		return nil, fmt.Errorf("get project	by id: %w", err)
	}
	return &project, nil
}
func (p *PostgresProjectRepository) List(ctx context.Context, userID int) ([]*domain.Project, error) {
	if userID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, stmtGetListProjects, userID)

	if err != nil {
		return nil, fmt.Errorf("query list project: %w", err)
	}

	defer rows.Close()

	projects := make([]*domain.Project, 0)
	for rows.Next() {
		var project domain.Project

		err := rows.Scan(&project.ID, &project.CreatedBy, &project.Title, &project.Goal, &project.Status, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		projects = append(projects, &project)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}
func (p *PostgresProjectRepository) Update(ctx context.Context, userID int, project *domain.Project) error {
	if userID <= 0 {
		return repository.ErrUserNotFound
	}
	if project == nil {
		return repository.ErrNilProject
	}
	if project.ID <= 0 {
		return repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := p.pool.Exec(ctx, stmtUpdateProject, project.Title, project.Goal, project.UpdatedAt, project.ID, userID)

	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrProjectNotFound
	}

	return nil
}
