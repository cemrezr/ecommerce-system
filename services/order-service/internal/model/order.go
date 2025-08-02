package model

import (
	"time"
)

type OrderRequest struct {
	UserID    int64 `json:"user_id" validate:"required,gt=0"`
	ProductID int64 `json:"product_id" validate:"required,gt=0"`
	Quantity  int   `json:"quantity" validate:"required,gt=0"`
}

func (r *OrderRequest) ToOrder() *Order {
	return &Order{
		UserID:    r.UserID,
		ProductID: r.ProductID,
		Quantity:  r.Quantity,
	}
}

type Order struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	ProductID int64     `db:"product_id" json:"product_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
