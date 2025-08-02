package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type EventLogger interface {
	Insert(ctx context.Context, log *EventLog) error
	UpdateStatus(ctx context.Context, id int64, status string, retryCount int) error
	ListFailed(ctx context.Context) ([]*EventLog, error)
}

type eventLogRepository struct {
	db *sqlx.DB
}

func NewEventLogRepository(db *sqlx.DB) EventLogger {
	return &eventLogRepository{db: db}
}

type EventLog struct {
	ID           int64     `db:"id"`
	EventType    string    `db:"event_type"`
	EventVersion string    `db:"event_version"`
	Payload      string    `db:"payload"`
	Status       string    `db:"status"`
	RetryCount   int       `db:"retry_count"`
	OrderID      *int64    `db:"order_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (r *eventLogRepository) Insert(ctx context.Context, e *EventLog) error {
	const q = `INSERT INTO event_logs (event_type, event_version, payload, status, retry_count, order_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, q,
		e.EventType, e.EventVersion, e.Payload,
		e.Status, e.RetryCount, e.OrderID).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *eventLogRepository) UpdateStatus(ctx context.Context, id int64, status string, retryCount int) error {
	const q = `UPDATE event_logs SET status = $1, retry_count = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, q, status, retryCount, id)
	return err
}

func (r *eventLogRepository) ListFailed(ctx context.Context) ([]*EventLog, error) {
	const q = `SELECT id, event_type, event_version, payload, status, retry_count, order_id, created_at, updated_at
	FROM event_logs WHERE status IN ('failed', 'publishing') ORDER BY created_at ASC`

	var list []*EventLog
	err := r.db.SelectContext(ctx, &list, q)
	return list, err
}
