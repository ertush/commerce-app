package handlers

import (
	"encoding/json"
	"net/http"

	"commerce-app/internal/auth"
	"commerce-app/internal/database"
	"commerce-app/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CustomerHandler struct {
	customerRepo *database.CustomerRepository
}

func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{
		customerRepo: &database.CustomerRepository{},
	}
}

// CreateCustomer creates a new customer
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.customerRepo.Create(&customer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(customer.ID, customer.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"customer": customer,
		"token":    token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetCustomer gets a customer by ID
func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	customer, err := h.customerRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// LoginCustomer handles customer login
func (h *CustomerHandler) LoginCustomer(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	customer, err := h.customerRepo.GetByEmail(loginRequest.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(customer.ID, customer.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"customer": customer,
		"token":    token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
