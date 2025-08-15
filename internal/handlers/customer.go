package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

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

	// Validate inputs
	if customer.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(customer.Email) {
		http.Error(w, "Email is invalid", http.StatusBadRequest)
		return
	}

	if customer.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if customer.Phone == "" {
		http.Error(w, "Phone is required", http.StatusBadRequest)
		return
	}

	if !regexp.MustCompile(`^[0-9]{10}$`).MatchString(customer.Phone) {
		http.Error(w, "Phone is invalid", http.StatusBadRequest)
		return
	}

	// Create customer
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
