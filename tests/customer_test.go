package tests

import (
	"testing"

	"commerce-app/internal/database"
	"commerce-app/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCustomerRepository_Create(t *testing.T) {

	// Initialize test database
	err := database.InitDB()
	assert.NoError(t, err)
	defer database.CloseDB()

	// Run migrations
	err = database.RunMigrations()
	assert.NoError(t, err)

	repo := &database.CustomerRepository{}

	// Test creating a customer
	customer := &models.Customer{
		Email: "test@example.com",
		Name:  "Test User",
		Phone: "1234567890",
	}

	err = repo.Create(customer)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, customer.ID)
	assert.NotZero(t, customer.CreatedAt)
	assert.NotZero(t, customer.UpdatedAt)

	//clean up
	repo.Delete(customer.ID)
}

func TestCustomerRepository_GetByID(t *testing.T) {

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
		Email: "test@example.com",
		Name:  "Test User",
		Phone: "1234567890",
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

	//clean up
	repo.Delete(customer.ID)
}

func TestCustomerRepository_GetByEmail(t *testing.T) {

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
		Email: "test@example.com",
		Name:  "Test User",
		Phone: "1234567890",
	}

	err = repo.Create(customer)
	assert.NoError(t, err)

	// Test getting the customer by email
	retrievedCustomer, err := repo.GetByEmail(customer.Email)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, retrievedCustomer.ID)
	assert.Equal(t, customer.Email, retrievedCustomer.Email)
	assert.Equal(t, customer.Name, retrievedCustomer.Name)
	assert.Equal(t, customer.Phone, retrievedCustomer.Phone)

	//clean up
	repo.Delete(customer.ID)
}

func TestCustomerRepository_Delete(t *testing.T) {

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
		Email: "test@example.com",
		Name:  "Test User",
		Phone: "1234567890",
	}

	err = repo.Create(customer)
	assert.NoError(t, err)

	// Test deleting the customer
	err = repo.Delete(customer.ID)
	assert.NoError(t, err)

	// Test getting the customer by ID after deletion
	retrievedCustomer, err := repo.GetByID(customer.ID)
	assert.Error(t, err)
	assert.Nil(t, retrievedCustomer)

	//clean up
	repo.Delete(customer.ID)
}
