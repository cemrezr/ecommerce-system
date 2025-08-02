package app

import (
	"net/http"

	"github.com/cemrezr/ecommerce-system/inventory-service/internal/handler"
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/repository"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

func startHTTPServer(port string, repo repository.ProductRepository, log zerolog.Logger) {
	handler := handler.NewProductHandler(repo, log)

	router := mux.NewRouter()
	router.HandleFunc("/products", handler.CreateProduct).Methods("POST")
	router.HandleFunc("/products", handler.ListProducts).Methods("GET")
	router.HandleFunc("/products/{product_id}", handler.UpdateProduct).Methods("PUT")
	router.HandleFunc("/products/{product_id}", handler.DeleteProduct).Methods("DELETE")
	router.HandleFunc("/products/{id}", handler.GetProduct).Methods("GET")

	go func() {
		log.Info().Str("addr", ":"+port).Msg("Starting HTTP server for product management")
		if err := http.ListenAndServe(":"+port, router); err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()
}
