package model

type Order struct {
	ID        int64 `json:"id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}
