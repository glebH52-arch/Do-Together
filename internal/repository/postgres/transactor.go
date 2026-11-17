package postgres

import (
	"context"
	"do-together/internal/repository"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.Transactor = (*PostgresTransactor)(nil)

type transactionContextKey struct{}

type PostgresTransactor struct {
	pool *pgxpool.Pool
}

type databaseExecutor interface {
	Exec(
		ctx context.Context,
		sql string,
		arguments ...any,
	) (pgconn.CommandTag, error)

	QueryRow(
		ctx context.Context,
		sql string,
		args ...any,
	) pgx.Row
}

func NewPostgresTransactor(pool *pgxpool.Pool) *PostgresTransactor {
	return &PostgresTransactor{
		pool: pool,
	}
}

func (t *PostgresTransactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(transactionContextKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txCtx := context.WithValue(
		ctx,
		transactionContextKey{},
		tx,
	)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
func executorFromContext(ctx context.Context, pool *pgxpool.Pool) databaseExecutor {
	if tx, ok := transactionFromContext(ctx); ok {
		return tx
	}

	return pool
}
func transactionFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(transactionContextKey{}).(pgx.Tx)
	return tx, ok
}
