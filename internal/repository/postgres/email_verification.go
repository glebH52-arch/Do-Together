package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.EmailVerificationCodeRepository = (*PostgresEmailVerificationCodeRepository)(nil)

type PostgresEmailVerificationCodeRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEmailVerificationCodeRepository(pool *pgxpool.Pool) *PostgresEmailVerificationCodeRepository {
	return &PostgresEmailVerificationCodeRepository{
		pool: pool,
	}
}

func (r *PostgresEmailVerificationCodeRepository) Create(ctx context.Context, code *domain.EmailVerificationCode) error {
	if code == nil {
		return repository.ErrEmailVerificationCodeNil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	executor := executorFromContext(ctx, r.pool)
	err := executor.QueryRow(
		ctx,
		stmtCreateVerificationCode,
		code.UserID,
		code.CodeHash,
		code.ExpiresAt,
		code.CreatedAt,
	).Scan(&code.ID)
	if err != nil {
		return fmt.Errorf("create email verification code: %w", err)
	}

	return nil
}

func (r *PostgresEmailVerificationCodeRepository) GetLatestForUpdate(ctx context.Context, userID uuid.UUID) (*domain.EmailVerificationCode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if userID == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}

	tx, ok := transactionFromContext(ctx)
	if !ok {
		return nil, repository.ErrTransactionRequired
	}

	var code domain.EmailVerificationCode

	err := tx.QueryRow(
		ctx,
		stmtGetLatestVerificationCodeForUpdate,
		userID,
	).Scan(
		&code.ID,
		&code.UserID,
		&code.CodeHash,
		&code.ExpiresAt,
		&code.ConsumedAt,
		&code.FailedAttempts,
		&code.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrEmailVerificationCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf(
			"get latest email verification code: %w",
			err,
		)
	}

	return &code, nil
}
func (r *PostgresEmailVerificationCodeRepository) UpdateFailedAttempts(ctx context.Context, code *domain.EmailVerificationCode) error {
	if code == nil {
		return repository.ErrEmailVerificationCodeNil
	}
	if err := ctx.Err(); err != nil {
		return nil
	}
	if code.ID == uuid.Nil() {
		return repository.ErrEmailVerificationCodeNotFound
	}

	tx, ok := transactionFromContext(ctx)
	if !ok {
		return repository.ErrTransactionRequired
	}

	commandTag, err := tx.Exec(
		ctx,
		stmtUpdateVerificationFailedAttempts,
		code.ID,
		code.FailedAttempts,
	)
	if err != nil {
		return fmt.Errorf(
			"update verification failed attempts: %w",
			err,
		)
	}

	if commandTag.RowsAffected() != 1 {
		return repository.ErrEmailVerificationCodeNotFound
	}

	return nil

}
func (r *PostgresEmailVerificationCodeRepository) Consume(ctx context.Context, code *domain.EmailVerificationCode) error {
	if code == nil {
		return repository.ErrEmailVerificationCodeNil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if code.ID == uuid.Nil() {
		return repository.ErrEmailVerificationCodeNotFound
	}
	if code.ConsumedAt == nil {
		return repository.ErrEmailVerificationCodeNotConsumed
	}

	tx, ok := transactionFromContext(ctx)
	if !ok {
		return repository.ErrTransactionRequired
	}
	commandTag, err := tx.Exec(ctx, stmtConsumeVerificationCode, code.ID, *code.ConsumedAt)
	if err != nil {
		return fmt.Errorf(
			"consume verification code failed: %w",
			err,
		)
	}

	if commandTag.RowsAffected() != 1 {
		return repository.ErrEmailVerificationCodeNotFound
	}

	return nil

}
