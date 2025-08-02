package client

import (
	"encoding/json"
	"fmt"
	"github.com/cemrezr/ecommerce-system/order-service/internal/utils"
	"github.com/sony/gobreaker"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type InventoryClient struct {
	BaseURL string
	Client  *http.Client
	log     zerolog.Logger
	breaker *gobreaker.CircuitBreaker
}

func NewInventoryClient(baseURL string, log zerolog.Logger) *InventoryClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "inventory-http",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})
	return &InventoryClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 5 * time.Second},
		log:     log,
		breaker: cb,
	}
}

type Product struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Stock       int    `json:"stock"`
}

func (c *InventoryClient) GetProductByID(productID int64) (*Product, error) {
	url := fmt.Sprintf("%s/products/%d", c.BaseURL, productID)

	var product *Product
	_, err := utils.RetryWithBreaker(c.breaker, func() error {
		c.log.Debug().Str("url", url).Msg("Requesting product from inventory-service")

		resp, err := c.Client.Get(url)
		if err != nil {
			c.log.Error().Err(err).Msg("HTTP request failed")
			return err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			var p Product
			if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
				c.log.Error().Err(err).Msg("Failed to decode response")
				return err
			}
			product = &p
			return nil
		case http.StatusNotFound:
			return fmt.Errorf("product not found")
		default:
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
	})

	if err != nil {
		c.log.Error().Err(err).Int64("product_id", productID).Msg("Final failure from inventory-service")
		return nil, err
	}

	c.log.Info().Int64("product_id", product.ProductID).Str("product_name", product.ProductName).Msg("Product fetched successfully")
	return product, nil
}
