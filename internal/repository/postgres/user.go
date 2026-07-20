package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.UserRepository = (*PostgresUserRepository)(nil)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		pool: pool,
	}

}

func (p *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return repository.ErrNilUser
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var pgErr *pgconn.PgError
	err := p.pool.QueryRow(ctx, stmtCreateUser, user.Username, user.Email, user.PasswordHash, user.Status, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return repository.ErrUsernameAlreadyExists
			case "users_email_key":
				return repository.ErrUserEmailAlreadyExists
			}
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (p *PostgresUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var user domain.User
	err := p.pool.QueryRow(ctx, stmtGetUserByID, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (p *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var user domain.User
	err := p.pool.QueryRow(ctx, stmtGetUserByEmail, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}
