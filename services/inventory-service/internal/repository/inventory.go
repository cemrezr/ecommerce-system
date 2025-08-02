package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type InventoryRepository interface {
	DecreaseStock(ctx context.Context, productID int64, quantity int) error
	LogStockChange(ctx context.Context, productID int64, change int, reason string) error
	IncreaseStock(ctx context.Context, productID int64, quantity int) error
}

type PostgresInventoryRepository struct {
	db *sqlx.DB
}

func NewPostgresInventoryRepository(db *sqlx.DB) *PostgresInventoryRepository {
	return &PostgresInventoryRepository{db: db}
}

func (r *PostgresInventoryRepository) DecreaseStock(ctx context.Context, productID int64, quantity int) error {
	query := `UPDATE inventory SET stock = stock - $1, updated_at = NOW() WHERE product_id = $2 AND stock >= $1`
	res, err := r.db.ExecContext(ctx, query, quantity, productID)
	if err != nil {
		return fmt.Errorf("failed to decrease stock: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("stock insufficient or product not found")
	}

	return nil
}

func (r *PostgresInventoryRepository) LogStockChange(ctx context.Context, productID int64, change int, reason string) error {
	query := `
		INSERT INTO stock_logs (product_id, change, reason)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, productID, change, reason)
	if err != nil {
		return fmt.Errorf("failed to insert stock log: %w", err)
	}
	return nil
}

func (r *PostgresInventoryRepository) IncreaseStock(ctx context.Context, productID int64, quantity int) error {
	query := `
		UPDATE inventory
		SET stock = stock + $1, updated_at = NOW()
		WHERE product_id = $2;
	`
	res, err := r.db.ExecContext(ctx, query, quantity, productID)
	if err != nil {
		return fmt.Errorf("increase stock failed: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("product not found to increase stock")
	}

	return nil
}
