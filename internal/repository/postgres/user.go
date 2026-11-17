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
	executor := executorFromContext(ctx, p.pool)
	var pgErr *pgconn.PgError
	err := executor.QueryRow(ctx, stmtCreateUser, user.Username, user.DisplayName, user.Email, user.PasswordHash, user.Status, user.CreatedAt).Scan(&user.ID)
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

func (p *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var user domain.User
	err := p.pool.QueryRow(ctx, stmtGetUserByID, id).Scan(&user.ID, &user.Username, &user.DisplayName, &user.AvatarFilename, &user.AvatarUpdatedAt, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.EmailVerifiedAt)
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

	executor := executorFromContext(ctx, p.pool)

	err := executor.QueryRow(ctx, stmtGetUserByEmail, email).Scan(&user.ID, &user.Username, &user.DisplayName, &user.AvatarFilename, &user.AvatarUpdatedAt, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.EmailVerifiedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (p *PostgresUserRepository) UpdateProfile(ctx context.Context, user *domain.User) error {
	if user == nil {
		return repository.ErrNilUser
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	result, err := p.pool.Exec(ctx, stmtUpdateUserProfile, user.ID, user.DisplayName, user.AvatarFilename, user.AvatarUpdatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	if result.RowsAffected() != 1 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (p *PostgresUserRepository) UpdateEmailVerification(ctx context.Context, user *domain.User) error {
	if user == nil {
		return repository.ErrNilUser
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if user.EmailVerifiedAt == nil {
		return repository.ErrUserEmailNotVerified
	}
	tx, ok := transactionFromContext(ctx)
	if !ok {
		return repository.ErrTransactionRequired
	}
	commandTag, err := tx.Exec(ctx, stmtUpdateUserEmailVerification, user.ID, *user.EmailVerifiedAt)
	if err != nil {
		return fmt.Errorf(
			"email verification failed: %w",
			err,
		)
	}

	if commandTag.RowsAffected() != 1 {
		return domain.ErrEmailAlreadyVerified
	}

	return nil
}
