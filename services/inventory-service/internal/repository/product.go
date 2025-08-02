package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, name string, stock int) (*Product, error)
	UpdateProduct(ctx context.Context, productID int64, name string, stock int) (*Product, error)
	DeleteProduct(ctx context.Context, productID int64) error
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id int64) (*Product, error)
}

type Product struct {
	ProductName string    `json:"product_name"`
	ProductID   int64     `json:"product_id"`
	Stock       int       `json:"stock"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) InsertProduct(ctx context.Context, name string, stock int) (*Product, error) {
	query := `
		INSERT INTO inventory (product_name, stock)
		VALUES ($1, $2)
		RETURNING product_id, product_name, stock, updated_at;
	`

	var product Product
	err := r.db.QueryRowContext(ctx, query, name, stock).Scan(
		&product.ProductID,
		&product.ProductName,
		&product.Stock,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert product: %w", err)
	}

	return &product, nil
}

func (r *PostgresProductRepository) UpdateProduct(ctx context.Context, productID int64, name string, stock int) (*Product, error) {
	query := `
		UPDATE inventory
		SET stock = $1, product_name = $2, updated_at = NOW()
		WHERE product_id = $3
		RETURNING product_id, product_name, stock, updated_at;
	`

	var product Product
	err := r.db.QueryRowContext(ctx, query, stock, name, productID).Scan(
		&product.ProductID,
		&product.ProductName,
		&product.Stock,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &product, nil
}

func (r *PostgresProductRepository) DeleteProduct(ctx context.Context, productID int64) error {
	query := `
		DELETE FROM inventory
		WHERE product_id = $1;
	`
	_, err := r.db.ExecContext(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("delete product failed: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) GetAllProducts(ctx context.Context) ([]Product, error) {
	query := `SELECT product_id, product_name, stock, updated_at FROM inventory ORDER BY product_id ASC;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all products failed: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.Stock, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id int64) (*Product, error) {
	query := `SELECT product_id, product_name, stock, updated_at FROM inventory WHERE product_id = $1`

	var p Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ProductID, &p.ProductName, &p.Stock, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get product by ID failed: %w", err)
	}
	return &p, nil
}
