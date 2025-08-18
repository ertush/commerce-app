package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"commerce-app/api/rest"
	"commerce-app/internal/auth"
	"commerce-app/internal/database"
	"commerce-app/internal/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
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
	if err := godotenv.Load(".env.test"); err != nil {
		log.Println("No .env file found")
	}

	os.Setenv("DB_HOST", os.Getenv("DB_HOST"))
	os.Setenv("DB_PORT", os.Getenv("DB_PORT"))
	os.Setenv("DB_USER", os.Getenv("DB_USER"))
	os.Setenv("DB_PASSWORD", os.Getenv("DB_PASSWORD"))
	os.Setenv("DB_NAME", os.Getenv("DB_NAME"))

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

func (ts *TestServer) GetTestCustomer(t *testing.T) (*models.Customer, string) {

	var customerData = GetTestUserDetails()
	var TestCustomer, token = CreateTestCustomer(t, ts, customerData)

	return TestCustomer, token
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
func getRandomEmail() string {
	return fmt.Sprintf("%s@example.com", uuid.New().String()[0:8])
}

func getRandomPhoneNumber() string {
	return fmt.Sprintf("%d", rand.Intn(9999999999))
}

func GetTestUserDetails() models.Customer {
	// Create test user
	userDetails := models.Customer{
		ID:    uuid.New(),
		Email: getRandomEmail(),
		Name:  "Test User",
		Phone: getRandomPhoneNumber(),
	}

	return userDetails

}

// CreateTestCustomer creates a test customer and returns the customer data
func CreateTestCustomer(t *testing.T, ts *TestServer, customerData models.Customer) (*models.Customer, string) {

	token, err := auth.GenerateToken(customerData.ID, customerData.Email)
	if err != nil {
		t.Fatal(err)
	}
	jsonData, _ := json.Marshal(customerData)
	req, _ := http.NewRequest("POST", ts.Server.URL+"/api/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	customer := response["customer"].(map[string]interface{})
	token = response["token"].(string)

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

	jsonData, err := json.Marshal(productData)
	if err != nil {
		t.Fatalf("Failed to marshal product data: %v", err)
	}

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
