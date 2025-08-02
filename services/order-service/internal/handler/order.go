package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cemrezr/ecommerce-system/order-service/internal/client"
	"github.com/cemrezr/ecommerce-system/order-service/internal/event"
	"github.com/cemrezr/ecommerce-system/order-service/internal/model"
	"github.com/cemrezr/ecommerce-system/order-service/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type OrderHandler struct {
	Repo            repository.OrderRepository
	Publisher       *event.Publisher
	validator       *validator.Validate
	InventoryClient *client.InventoryClient
	log             zerolog.Logger
}

func NewOrderHandler(
	repo repository.OrderRepository,
	pub *event.Publisher,
	invClient *client.InventoryClient,
	log zerolog.Logger,
) *OrderHandler {
	return &OrderHandler{
		Repo:            repo,
		Publisher:       pub,
		validator:       validator.New(),
		InventoryClient: invClient,
		log:             log,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn().Err(err).Msg("Invalid JSON payload")
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.log.Warn().Err(err).Msg("Validation failed")
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, e := range validationErrs {
				field := e.Field()
				switch e.Tag() {
				case "required":
					errors[field] = field + " is required"
				case "gt":
					errors[field] = field + " must be greater than " + e.Param()
				default:
					errors[field] = "Invalid value for " + field
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Validation failed",
				"errors":  errors,
			})
			return
		}
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	product, err := h.InventoryClient.GetProductByID(req.ProductID)
	if err != nil {
		h.log.Warn().Err(err).Int64("product_id", req.ProductID).Msg("Product lookup failed")
		http.Error(w, "Product not found", http.StatusBadRequest)
		return
	}

	if product.Stock < req.Quantity {
		h.log.Warn().
			Int64("product_id", req.ProductID).
			Int("stock", product.Stock).
			Int("requested", req.Quantity).
			Msg("Insufficient stock")
		http.Error(w, "Insufficient stock", http.StatusBadRequest)
		return
	}

	order := req.ToOrder()

	if err := h.Repo.Create(r.Context(), order); err != nil {
		h.log.Error().Err(err).Msg("Failed to create order in DB")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.log.Info().
		Int64("order_id", order.ID).
		Int64("user_id", order.UserID).
		Int64("product_id", order.ProductID).
		Int("quantity", order.Quantity).
		Msg("✅ Order successfully created")

	if err := h.Publisher.PublishOrderCreated(order); err != nil {
		h.log.Error().
			Err(err).
			Str("event", "order.created").
			Int64("order_id", order.ID).
			Msg("Failed to publish order.created event")
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["order_id"]
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warn().Str("order_id", idStr).Msg("Invalid order_id format")
		http.Error(w, "Invalid order_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	order, err := h.Repo.GetByID(ctx, orderID)
	if err != nil {
		h.log.Error().Err(err).Int64("order_id", orderID).Msg("Order not found")
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	if err := h.Repo.Cancel(ctx, orderID); err != nil {
		h.log.Error().Err(err).Int64("order_id", orderID).Msg("Failed to cancel order")
		http.Error(w, "Cancel failed", http.StatusInternalServerError)
		return
	}

	if err := h.Publisher.PublishOrderCancelled(order); err != nil {
		h.log.Error().Err(err).Int64("order_id", orderID).Msg("Failed to publish order.cancelled event")
	}

	h.log.Info().Int64("order_id", orderID).Msg("✅ Order cancelled and event published")
	w.WriteHeader(http.StatusOK)
}
