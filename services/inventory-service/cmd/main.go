package main

import (
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/app"
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/config"
	"github.com/cemrezr/ecommerce-system/pkg/logger"
)

func main() {
	log := logger.NewLogger("inventory-service")
	cfg := config.Load()
	app.Run(cfg, log)
}
