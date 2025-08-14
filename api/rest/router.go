package rest

import (
	"log"
	"net/http"
	"os"

	"commerce-app/internal/auth"
	"commerce-app/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Router sets up the REST API routes
func Router() http.Handler {
	r := mux.NewRouter()

	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler()
	productHandler := handlers.NewProductHandler()
	orderHandler := handlers.NewOrderHandler()

	// Initialize auth handler
	authHandler, err := handlers.NewAuthHandler()
	if err != nil {
		log.Printf("Warning: Failed to initialize OIDC auth handler: %v", err)
		authHandler = nil
	}

	// Initialize OIDC middleware
	var oidcMiddleware *auth.OIDCMiddleware
	if authHandler != nil {
		jwtSecret := []byte(getEnv("JWT_SECRET", "secret"))
		oidcMiddleware = auth.NewOIDCMiddleware(authHandler.GetOIDCProvider(), jwtSecret)
	}

	// OIDC Authentication routes
	if authHandler != nil {
		r.HandleFunc("/api/auth/login", authHandler.Login).Methods("GET")
		r.HandleFunc("/api/auth/callback", authHandler.Callback).Methods("GET")
		r.HandleFunc("/api/auth/logout", authHandler.Logout).Methods("POST")
	}

	// Protected customer routes (require authentication)
	protectedCustomer := r.PathPrefix("/api/customers").Subrouter()
	if oidcMiddleware != nil {
		protectedCustomer.Use(oidcMiddleware.RequireAuth)
	}

	protectedCustomer.HandleFunc("", customerHandler.CreateCustomer).Methods("POST")
	protectedCustomer.HandleFunc("/{id}", customerHandler.GetCustomer).Methods("GET")
	protectedCustomer.HandleFunc("/{customerId}/orders", orderHandler.GetOrdersByCustomer).Methods("GET")

	// Protected user info route (require authentication)
	protectedUserInfo := r.PathPrefix("/api/auth/userinfo").Subrouter()
	if oidcMiddleware != nil {
		protectedUserInfo.Use(oidcMiddleware.RequireAuth)
	}

	// Product routes
	r.HandleFunc("/api/products", productHandler.CreateProduct).Methods("POST")
	r.HandleFunc("/api/products", productHandler.GetAllProducts).Methods("GET")
	r.HandleFunc("/api/products/{id}", productHandler.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/category/{categoryId}", productHandler.GetProductsByCategory).Methods("GET")
	r.HandleFunc("/api/products/category/{categoryId}/average-price", productHandler.GetAveragePriceByCategory).Methods("GET")

	// Category routes
	r.HandleFunc("/api/categories", productHandler.CreateCategory).Methods("POST")
	r.HandleFunc("/api/categories", productHandler.GetAllCategories).Methods("GET")
	r.HandleFunc("/api/categories/{id}", productHandler.GetCategory).Methods("GET")
	r.HandleFunc("/api/categories/{parentId}/children", productHandler.GetCategoryChildren).Methods("GET")

	// Order routes
	r.HandleFunc("/api/orders", orderHandler.CreateOrder).Methods("POST")
	r.HandleFunc("/api/orders/{id}", orderHandler.GetOrder).Methods("GET")
	r.HandleFunc("/api/orders/{id}/status", orderHandler.UpdateOrderStatus).Methods("PUT")

	// Protected routes (require authentication)
	// protected := r.PathPrefix("/api/protected").Subrouter()
	// if oidcMiddleware != nil {
	// 	protected.Use(oidcMiddleware.RequireAuth)
	// } else {
	// 	protected.Use(auth.AuthMiddleware)
	// }

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	return c.Handler(r)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
