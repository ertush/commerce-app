package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer in the system
type Customer struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Category represents a product category with hierarchical structure
type Category struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	ParentID    *uuid.UUID `json:"parent_id" db:"parent_id"`
	Level       int        `json:"level" db:"level"`
	Path        string     `json:"path" db:"path"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	Children    []Category `json:"children,omitempty"`
}

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`
	Stock       int       `json:"stock" db:"stock"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Category    Category  `json:"category"`
}

// Order represents an order in the system
type Order struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	CustomerID uuid.UUID   `json:"customer_id" db:"customer_id"`
	Status     string      `json:"status" db:"status"`
	Total      float64     `json:"total" db:"total"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
	Customer   Customer    `json:"customer"`
	Items      []OrderItem `json:"items,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrderID   uuid.UUID `json:"order_id" db:"order_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	Product   Product   `json:"product"`
}

// CategoryPrice represents average price for a category
type CategoryPrice struct {
	CategoryID   uuid.UUID `json:"category_id" db:"category_id"`
	CategoryName string    `json:"category_name" db:"category_name"`
	AveragePrice float64   `json:"average_price" db:"average_price"`
	ProductCount int       `json:"product_count" db:"product_count"`
}
