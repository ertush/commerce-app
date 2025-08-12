package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"commerce-app/api/rest"
	"commerce-app/internal/database"
	"commerce-app/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestServer represents a test server instance
type TestServer struct {
	Server *httptest.Server
	DB     *sql.DB
}

// SetupTestServer creates a test server with a clean database
func SetupTestServer(t *testing.T) *TestServer {
	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_NAME", "ecommerce_test")

	// Initialize database
	err := database.InitDB()
	assert.NoError(t, err)

	// Run migrations
	err = database.RunMigrations()
	assert.NoError(t, err)

	// Create router
	router := rest.Router()

	// Create test server
	server := httptest.NewServer(router)

	return &TestServer{
		Server: server,
		DB:     database.DB,
	}
}

// CleanupTestServer cleans up the test server and database
func (ts *TestServer) CleanupTestServer(t *testing.T) {
	if ts.Server != nil {
		ts.Server.Close()
	}
	if ts.DB != nil {
		database.CloseDB()
	}
}

// CreateTestCustomer creates a test customer and returns the customer data
func CreateTestCustomer(t *testing.T, ts *TestServer) (*models.Customer, string) {
	customerData := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
		"phone": "1234567890",
	}

	jsonData, _ := json.Marshal(customerData)
	req, _ := http.NewRequest("POST", ts.Server.URL+"/api/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	customer := response["customer"].(map[string]interface{})
	token := response["token"].(string)

	return &models.Customer{
		ID:    uuid.MustParse(customer["id"].(string)),
		Email: customer["email"].(string),
		Name:  customer["name"].(string),
		Phone: customer["phone"].(string),
	}, token
}

// CreateTestCategory creates a test category and returns the category data
func CreateTestCategory(t *testing.T, ts *TestServer, parentID *uuid.UUID) *models.Category {
	categoryData := map[string]interface{}{
		"name":        "Test Category",
		"description": "Test category description",
	}
	if parentID != nil {
		categoryData["parent_id"] = parentID.String()
	}

	jsonData, _ := json.Marshal(categoryData)
	req, _ := http.NewRequest("POST", ts.Server.URL+"/api/categories", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var category models.Category
	err = json.NewDecoder(resp.Body).Decode(&category)
	assert.NoError(t, err)

	return &category
}

// CreateTestProduct creates a test product and returns the product data
func CreateTestProduct(t *testing.T, ts *TestServer, categoryID uuid.UUID) *models.Product {
	productData := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test product description",
		"price":       99.99,
		"category_id": categoryID.String(),
		"stock":       10,
		"image_url":   "https://example.com/image.jpg",
	}

	jsonData, _ := json.Marshal(productData)
	req, _ := http.NewRequest("POST", ts.Server.URL+"/api/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var product models.Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	assert.NoError(t, err)

	return &product
}

// CreateTestOrder creates a test order and returns the order data
func CreateTestOrder(t *testing.T, ts *TestServer, customerID uuid.UUID, productID uuid.UUID) *models.Order {
	orderData := map[string]interface{}{
		"customer_id": customerID.String(),
		"items": []map[string]interface{}{
			{
				"product_id": productID.String(),
				"quantity":   2,
			},
		},
	}

	jsonData, _ := json.Marshal(orderData)
	req, _ := http.NewRequest("POST", ts.Server.URL+"/api/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var order models.Order
	err = json.NewDecoder(resp.Body).Decode(&order)
	assert.NoError(t, err)

	return &order
}

// MakeRequest is a helper function to make HTTP requests
func MakeRequest(t *testing.T, ts *TestServer, method, path string, body interface{}, headers map[string]string) *http.Response {
	var jsonData []byte
	var err error

	if body != nil {
		jsonData, err = json.Marshal(body)
		assert.NoError(t, err)
	}

	req, err := http.NewRequest(method, ts.Server.URL+path, bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	return resp
}
