package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cemrezr/ecommerce-system/inventory-service/internal/repository"
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/utils"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type ProductHandler struct {
	Repo repository.ProductRepository
	Log  zerolog.Logger
}

func NewProductHandler(repo repository.ProductRepository, log zerolog.Logger) *ProductHandler {
	return &ProductHandler{Repo: repo, Log: log}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductName string `json:"product_name"`
		Stock       int    `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.Warn().Err(err).Msg("Invalid product creation payload")
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	product, err := h.Repo.InsertProduct(r.Context(), req.ProductName, req.Stock)
	if err != nil {
		h.Log.Error().Err(err).Str("product_name", req.ProductName).Msg("Failed to insert product")
		utils.WriteError(w, http.StatusInternalServerError, "Failed to insert product")
		return
	}

	h.Log.Info().Int64("product_id", product.ProductID).Msg("Product created")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["product_id"]
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.Log.Warn().Str("product_id", idStr).Msg("Invalid product_id path param")
		utils.WriteError(w, http.StatusBadRequest, "Invalid product_id path param")
		return
	}

	var req struct {
		ProductName string `json:"product_name"`
		Stock       int    `json:"stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.Warn().Err(err).Msg("Invalid update product payload")
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	product, err := h.Repo.UpdateProduct(r.Context(), productID, req.ProductName, req.Stock)
	if err != nil {
		h.Log.Error().Err(err).Int64("product_id", productID).Msg("Failed to update product")
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	h.Log.Info().Int64("product_id", product.ProductID).Msg("Product updated")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["product_id"]
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.Log.Warn().Str("product_id", idStr).Msg("Invalid product_id path param")
		utils.WriteError(w, http.StatusBadRequest, "Invalid product_id path param")
		return
	}

	if err := h.Repo.DeleteProduct(r.Context(), productID); err != nil {
		h.Log.Error().Err(err).Int64("product_id", productID).Msg("Failed to delete product")
		utils.WriteError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	h.Log.Info().Int64("product_id", productID).Msg("Product deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.Repo.GetAllProducts(context.Background())
	if err != nil {
		h.Log.Error().Err(err).Msg("Failed to list products")
		utils.WriteError(w, http.StatusInternalServerError, "Failed to list products")
		return
	}

	h.Log.Info().Int("count", len(products)).Msg("Product list retrieved")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.Log.Warn().Str("id", idStr).Msg("Invalid product id")
		utils.WriteError(w, http.StatusBadRequest, "Invalid product id")
		return
	}

	product, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		h.Log.Error().Err(err).Int64("product_id", id).Msg("Product not found")
		utils.WriteError(w, http.StatusNotFound, "Product not found")
		return
	}

	h.Log.Info().Int64("product_id", product.ProductID).Msg("Product fetched")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
