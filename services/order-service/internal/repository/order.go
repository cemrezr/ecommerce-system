package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cemrezr/ecommerce-system/order-service/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	Cancel(ctx context.Context, orderID int64) error
	GetByID(ctx context.Context, id int64) (*model.Order, error)
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	order.CreatedAt = time.Now()
	order.Status = "created"

	query := `
		INSERT INTO orders (user_id, product_id, quantity, status, created_at)
		VALUES (:user_id, :product_id, :quantity, :status, :created_at)
		RETURNING id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Str("operation", "PrepareNamedContext").Msg("Failed to prepare insert statement")
		return err
	}

	err = stmt.GetContext(ctx, &order.ID, order)
	if err != nil {
		log.Error().Err(err).Str("operation", "ExecInsert").Int64("user_id", order.UserID).Int64("product_id", order.ProductID).Msg("Failed to execute order insert")
		return err
	}

	log.Info().Str("operation", "CreateOrder").Int64("order_id", order.ID).Int64("user_id", order.UserID).Int64("product_id", order.ProductID).Msg("✅ Order inserted into database")
	return nil
}

func (r *orderRepository) Cancel(ctx context.Context, orderID int64) error {
	query := `UPDATE orders SET status = 'cancelled', updated_at = NOW() WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("order not found or already cancelled")
	}

	return nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var order model.Order
	query := `SELECT id, user_id, product_id, quantity FROM orders WHERE id = $1`

	err := r.db.GetContext(ctx, &order, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	return &order, nil
}
