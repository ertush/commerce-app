package tests

import (
	"testing"

	"commerce-app/internal/database"
	"commerce-app/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCustomerRepository_Create(t *testing.T) {
	ts := SetupTestServer(t)

	// Initialize test database
	err := database.InitDB()
	assert.NoError(t, err)
	defer database.CloseDB()

	// Run migrations
	err = database.RunMigrations()
	assert.NoError(t, err)

	repo := &database.CustomerRepository{}

	// Test creating a customer
	customer, _ := ts.GetTestCustomer(t)

	err = repo.Create(customer)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, customer.ID)
	assert.NotZero(t, customer.CreatedAt)
	assert.NotZero(t, customer.UpdatedAt)
}

func TestCustomerRepository_GetByID(t *testing.T) {
	ts := SetupTestServer(t)

	// Initialize test database
	err := database.InitDB()
	assert.NoError(t, err)
	defer database.CloseDB()

	// Run migrations
	err = database.RunMigrations()
	assert.NoError(t, err)

	repo := &database.CustomerRepository{}

	// Create a customer first
	customer := &models.Customer{
		Email: "test2@example.com",
		Name:  "Test User 2",
		Phone: "1234567891",
	}

	err = repo.Create(customer)
	assert.NoError(t, err)

	// Test getting the customer by ID
	retrievedCustomer, err := repo.GetByID(customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, retrievedCustomer.ID)
	assert.Equal(t, customer.Email, retrievedCustomer.Email)
	assert.Equal(t, customer.Name, retrievedCustomer.Name)
	assert.Equal(t, customer.Phone, retrievedCustomer.Phone)
}

func TestCustomerRepository_GetByEmail(t *testing.T) {
	ts := SetupTestServer(t)

	// Initialize test database
	err := database.InitDB()
	assert.NoError(t, err)
	defer database.CloseDB()

	// Run migrations
	err = database.RunMigrations()
	assert.NoError(t, err)

	repo := &database.CustomerRepository{}

	// Create a customer first
	customer, _ := ts.GetTestCustomer(t)

	err = repo.Create(customer)
	assert.NoError(t, err)

	// Test getting the customer by email
	retrievedCustomer, err := repo.GetByEmail(customer.Email)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, retrievedCustomer.ID)
	assert.Equal(t, customer.Email, retrievedCustomer.Email)
	assert.Equal(t, customer.Name, retrievedCustomer.Name)
	assert.Equal(t, customer.Phone, retrievedCustomer.Phone)
}
