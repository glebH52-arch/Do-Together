package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.InviteRepository = (*PostgresInviteRepository)(nil)

type PostgresInviteRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresInviteRepository(pool *pgxpool.Pool) *PostgresInviteRepository {
	return &PostgresInviteRepository{
		pool: pool,
	}
}

func (p *PostgresInviteRepository) Create(
	ctx context.Context,
	invite *domain.Invite,
) error {
	if invite == nil {
		return domain.ErrInviteNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin invite: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, stmtExpireStaleInvites, invite.ProjectID, invite.InviteeID, invite.CreatedAt); err != nil {
		return fmt.Errorf("expire stale invites: %w", err)
	}
	err = tx.QueryRow(ctx, stmtCreateInvite, invite.ProjectID, invite.InviterID, invite.InviteeID, invite.Role, invite.Status, invite.ExpiresAt, invite.CreatedAt).Scan(&invite.ID, &invite.ProjectTitle, &invite.InviterName)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			var canInvite, isMember bool
			if checkErr := tx.QueryRow(ctx, stmtCanInvite, invite.ProjectID, invite.InviterID).Scan(&canInvite); checkErr != nil {
				return checkErr
			}
			if !canInvite {
				return repository.ErrForbidden
			}
			if checkErr := tx.QueryRow(ctx, stmtIsProjectMember, invite.ProjectID, invite.InviteeID).Scan(&isMember); checkErr != nil {
				return checkErr
			}
			if isMember {
				return repository.ErrProjectMemberAlreadyExists
			}
			return repository.ErrForbidden
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			switch pgErr.Code {

			case "23505":
				if pgErr.ConstraintName == "invites_one_pending_idx" {
					return domain.ErrPendingInviteAlreadyExists
				}

			case "23503":
				switch pgErr.ConstraintName {

				case "invites_project_id_fkey":
					return repository.ErrProjectNotFound

				case "invites_inviter_id_fkey",
					"invites_invitee_id_fkey":
					return repository.ErrUserNotFound
				}
			}
		}

		return fmt.Errorf("insert invite: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit invite: %w", err)
	}
	return nil
}

func (p *PostgresInviteRepository) FindInviteeIDByEmail(ctx context.Context, email string) (int, error) {
	var id int
	err := p.pool.QueryRow(ctx, stmtFindInviteeByEmail, email).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, repository.ErrUserNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("find invitee: %w", err)
	}
	return id, nil
}

func (p *PostgresInviteRepository) List(
	ctx context.Context,
	inviteeID int,
) ([]*domain.Invite, error) {
	if inviteeID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, stmtGetListInvites, inviteeID)

	if err != nil {
		return nil, fmt.Errorf("query invites: %w", err)
	}

	defer rows.Close()

	invites := make([]*domain.Invite, 0)
	for rows.Next() {
		var invite domain.Invite

		err := rows.Scan(&invite.ID, &invite.ProjectID, &invite.InviterID, &invite.InviteeID, &invite.Role, &invite.Status, &invite.ExpiresAt, &invite.CreatedAt, &invite.UpdatedAt, &invite.ProjectTitle, &invite.InviterName)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		invites = append(invites, &invite)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invites: %w", err)
	}

	return invites, nil
}

func (p *PostgresInviteRepository) Accept(
	ctx context.Context,
	inviteeID, inviteID int,
	now time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin accept invite: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	var invite domain.Invite

	err = tx.QueryRow(ctx, stmtGetInviteForUpdate, inviteID, inviteeID).Scan(&invite.ID, &invite.ProjectID, &invite.InviterID, &invite.InviteeID, &invite.Role, &invite.Status, &invite.ExpiresAt, &invite.CreatedAt, &invite.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInviteNotFound
		}
		return fmt.Errorf("get invite for update: %w", err)
	}
	if err := invite.Accept(now); err != nil {
		return err
	}
	commandTag, err := tx.Exec(ctx, stmtCreateProjectMember, invite.ProjectID, invite.InviteeID, invite.Role, now)
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

	commandTag, err = tx.Exec(ctx, stmtAcceptInvite, invite.Status, now, invite.ID)
	if err != nil {
		return fmt.Errorf("mark invite accepted: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.ErrInviteNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit accept invite transaction: %w", err)
	}
	return nil
}

func (p *PostgresInviteRepository) Decline(
	ctx context.Context,
	inviteeID, inviteID int,
	now time.Time,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := p.pool.Exec(ctx, stmtDeclineInvite, inviteID, inviteeID, now)
	if err != nil {
		return fmt.Errorf("decline invite: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.ErrInviteNotFound
	}
	return nil
}
