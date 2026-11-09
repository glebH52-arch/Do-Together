package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.NotificationRepository = (*PostgresNotificationRepository)(nil)

type PostgresNotificationRepository struct{ pool *pgxpool.Pool }

func NewNotificationRepository(pool *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{pool: pool}
}

func (r *PostgresNotificationRepository) List(ctx context.Context, userID, limit, offset int) ([]*domain.Notification, error) {
	rows, err := r.pool.Query(ctx, sqlListNotifications, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		item := &domain.Notification{}
		if err := rows.Scan(&item.ID, &item.RecipientID, &item.ProjectID, &item.TaskID, &item.InviteID,
			&item.Kind, &item.Payload, &item.CreatedAt, &item.ReadAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, item)
	}
	return notifications, rows.Err()
}

func (r *PostgresNotificationRepository) MarkRead(ctx context.Context, userID int, notificationID int64) error {
	tag, err := r.pool.Exec(ctx, sqlMarkNotificationRead, notificationID, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotificationNotFound
	}
	return nil
}
