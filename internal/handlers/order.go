package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"commerce-app/internal/database"
	"commerce-app/internal/models"
	"commerce-app/internal/notifications"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderRepo    *database.OrderRepository
	customerRepo *database.CustomerRepository
	productRepo  *database.ProductRepository
	smsService   *notifications.SMSService
	emailService *notifications.EmailService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderRepo:    &database.OrderRepository{},
		customerRepo: &database.CustomerRepository{},
		productRepo:  &database.ProductRepository{},
		smsService:   notifications.NewSMSService(),
		emailService: notifications.NewEmailService(),
	}
}

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderRequest struct {
		CustomerID uuid.UUID `json:"customer_id"`
		Items      []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get customer
	customer, err := h.customerRepo.GetByID(orderRequest.CustomerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	// Create order
	order := &models.Order{
		CustomerID: orderRequest.CustomerID,
		Status:     "pending",
		Total:      0,
		Items:      []models.OrderItem{},
	}

	// Process items and calculate total
	var itemDetails []string
	for _, item := range orderRequest.Items {
		product, err := h.productRepo.GetByID(item.ProductID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Product not found: %s", item.ProductID), http.StatusNotFound)
			return
		}

		if product.Stock < item.Quantity {
			http.Error(w, fmt.Sprintf("Insufficient stock for product: %s", product.Name), http.StatusBadRequest)
			return
		}

		itemTotal := product.Price * float64(item.Quantity)
		order.Total += itemTotal

		orderItem := models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Product:   *product,
		}
		order.Items = append(order.Items, orderItem)

		itemDetails = append(itemDetails, fmt.Sprintf("- %s x%d @ Ksh%.2f = Ksh%.2f",
			product.Name, item.Quantity, product.Price, itemTotal))
	}

	// Create order in database
	if err := h.orderRepo.Create(order); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msgOrderID := order.ID.String()[0:8] + "..." + order.ID.String()[len(order.ID.String())-4:]

	// Send SMS notification to customer
	go func() {
		// "New Order Placed has been received.Order ID: #%s,\ncustomer name: %s,\nphone: %s,\nemail: %s,\ntotal: $%.2f,\nproduct: %s,\nqty: x%d",

		message := fmt.Sprintf(`
								New order has been Received!

								Order Details:
								- Order ID: %s
								- Name: %s
								- Phone: %s
								- Total Amount: Ksh%.2f

								Order Items:
								%s
								`,
			msgOrderID,
			customer.Name,
			customer.Phone,
			order.Total,
			strings.Join(itemDetails, "\n"))

		err := h.smsService.SendOrderNotification(message)
		if err != nil {
			fmt.Printf("Failed to send SMS notification: %v\n", err)
		}

	}()

	// Send email notification to admin
	go func() {
		if err := h.emailService.SendOrderNotificationToAdmin(
			msgOrderID,
			customer.Name,
			customer.Email,
			customer.Phone,
			order.Total,
			itemDetails,
		); err != nil {
			fmt.Printf("Failed to send email notification: %v\n", err)
		}
	}()

	orderTotal, err := strconv.ParseFloat(fmt.Sprintf("%.2f", order.Total), 64)
	if err != nil {
		fmt.Println("Error converting string to float64:", err)
		return
	}

	response := &models.OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     order.Status,
		Total:      orderTotal,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		Items:      order.Items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetOrder gets an order by ID
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetOrdersByCustomer gets orders for a customer
func (h *OrderHandler) GetOrdersByCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["customerId"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	orders, err := h.orderRepo.GetByCustomer(customerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// UpdateOrderStatus updates the status of an order
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	isValid := false
	for _, status := range validStatuses {
		if status == statusUpdate.Status {
			isValid = true
			break
		}
	}

	if !isValid {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Update order status (this would need to be implemented in the repository)
	// For now, we'll just return the order
	order, err := h.orderRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	order.Status = statusUpdate.Status

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
