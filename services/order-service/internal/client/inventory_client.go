package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type InventoryClient struct {
	BaseURL string
	Client  *http.Client
	log     zerolog.Logger
}

func NewInventoryClient(baseURL string, log zerolog.Logger) *InventoryClient {
	return &InventoryClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 5 * time.Second},
		log:     log,
	}
}

type Product struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Stock       int    `json:"stock"`
}

func (c *InventoryClient) GetProductByID(productID int64) (*Product, error) {
	url := fmt.Sprintf("%s/products/%d", c.BaseURL, productID)

	c.log.Debug().Str("url", url).Msg("Requesting product from inventory-service")

	resp, err := c.Client.Get(url)
	if err != nil {
		c.log.Error().Err(err).Int64("product_id", productID).Msg("Request to inventory-service failed")
		return nil, fmt.Errorf("request to inventory-service failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var product Product
		if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
			c.log.Error().Err(err).Int64("product_id", productID).Msg("Failed to decode inventory-service response")
			return nil, fmt.Errorf("failed to decode inventory-service response: %w", err)
		}
		c.log.Info().Int64("product_id", productID).Str("product_name", product.ProductName).Msg("Product retrieved from inventory-service")
		return &product, nil

	case http.StatusNotFound:
		c.log.Warn().Int64("product_id", productID).Msg("Product not found in inventory-service")
		return nil, fmt.Errorf("product not found")

	default:
		c.log.Error().
			Int64("product_id", productID).
			Int("status", resp.StatusCode).
			Msg("Unexpected status from inventory-service")
		return nil, fmt.Errorf("unexpected status from inventory-service: %d", resp.StatusCode)
	}
}
