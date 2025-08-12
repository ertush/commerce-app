package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"commerce-app/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestCustomerRoutes tests all customer-related endpoints
func TestCustomerRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateCustomer", func(t *testing.T) {
		customerData := map[string]interface{}{
			"email": "john@example.com",
			"name":  "John Doe",
			"phone": "1234567890",
		}

		resp := MakeRequest(t, ts, "POST", "/api/customers", customerData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Contains(t, response, "customer")
		assert.Contains(t, response, "token")

		customer := response["customer"].(map[string]interface{})
		assert.Equal(t, "john@example.com", customer["email"])
		assert.Equal(t, "John Doe", customer["name"])
		assert.Equal(t, "1234567890", customer["phone"])
	})

	t.Run("CreateCustomerInvalidData", func(t *testing.T) {
		customerData := map[string]interface{}{
			"email": "invalid-email",
			"name":  "",
			"phone": "",
		}

		resp := MakeRequest(t, ts, "POST", "/api/customers", customerData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("LoginCustomer", func(t *testing.T) {
		// First create a customer
		customer, _ := CreateTestCustomer(t, ts)

		loginData := map[string]interface{}{
			"email": customer.Email,
		}

		resp := MakeRequest(t, ts, "POST", "/api/customers/login", loginData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Contains(t, response, "customer")
		assert.Contains(t, response, "token")
	})

	t.Run("LoginCustomerNotFound", func(t *testing.T) {
		loginData := map[string]interface{}{
			"email": "nonexistent@example.com",
		}

		resp := MakeRequest(t, ts, "POST", "/api/customers/login", loginData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GetCustomer", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)

		resp := MakeRequest(t, ts, "GET", "/api/customers/"+customer.ID.String(), nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedCustomer models.Customer
		err := json.NewDecoder(resp.Body).Decode(&retrievedCustomer)
		assert.NoError(t, err)

		assert.Equal(t, customer.ID, retrievedCustomer.ID)
		assert.Equal(t, customer.Email, retrievedCustomer.Email)
		assert.Equal(t, customer.Name, retrievedCustomer.Name)
		assert.Equal(t, customer.Phone, retrievedCustomer.Phone)
	})

	t.Run("GetCustomerNotFound", func(t *testing.T) {
		randomID := uuid.New()
		resp := MakeRequest(t, ts, "GET", "/api/customers/"+randomID.String(), nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestCategoryRoutes tests all category-related endpoints
func TestCategoryRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateCategory", func(t *testing.T) {
		categoryData := map[string]interface{}{
			"name":        "Electronics",
			"description": "Electronic devices and accessories",
		}

		resp := MakeRequest(t, ts, "POST", "/api/categories", categoryData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var category models.Category
		err := json.NewDecoder(resp.Body).Decode(&category)
		assert.NoError(t, err)

		assert.Equal(t, "Electronics", category.Name)
		assert.Equal(t, "Electronic devices and accessories", category.Description)
		assert.Equal(t, 0, category.Level)
		assert.Equal(t, "/Electronics", category.Path)
	})

	t.Run("CreateSubCategory", func(t *testing.T) {
		parentCategory := CreateTestCategory(t, ts, nil)

		categoryData := map[string]interface{}{
			"name":        "Smartphones",
			"description": "Mobile phones and smartphones",
			"parent_id":   parentCategory.ID.String(),
		}

		resp := MakeRequest(t, ts, "POST", "/api/categories", categoryData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var category models.Category
		err := json.NewDecoder(resp.Body).Decode(&category)
		assert.NoError(t, err)

		assert.Equal(t, "Smartphones", category.Name)
		assert.Equal(t, 1, category.Level)
		assert.Equal(t, parentCategory.Path+"/Smartphones", category.Path)
	})

	t.Run("GetAllCategories", func(t *testing.T) {
		CreateTestCategory(t, ts, nil)

		resp := MakeRequest(t, ts, "GET", "/api/categories", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var categories []models.Category
		err := json.NewDecoder(resp.Body).Decode(&categories)
		assert.NoError(t, err)

		assert.Greater(t, len(categories), 0)
	})

	t.Run("GetCategory", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)

		resp := MakeRequest(t, ts, "GET", "/api/categories/"+category.ID.String(), nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedCategory models.Category
		err := json.NewDecoder(resp.Body).Decode(&retrievedCategory)
		assert.NoError(t, err)

		assert.Equal(t, category.ID, retrievedCategory.ID)
		assert.Equal(t, category.Name, retrievedCategory.Name)
	})

	t.Run("GetCategoryChildren", func(t *testing.T) {
		parentCategory := CreateTestCategory(t, ts, nil)
		CreateTestCategory(t, ts, &parentCategory.ID)

		resp := MakeRequest(t, ts, "GET", "/api/categories/"+parentCategory.ID.String()+"/children", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var children []models.Category
		err := json.NewDecoder(resp.Body).Decode(&children)
		assert.NoError(t, err)

		assert.Greater(t, len(children), 0)
	})
}

// TestProductRoutes tests all product-related endpoints
func TestProductRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateProduct", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)

		productData := map[string]interface{}{
			"name":        "iPhone 15",
			"description": "Latest iPhone model",
			"price":       999.99,
			"category_id": category.ID.String(),
			"stock":       10,
			"image_url":   "https://example.com/iphone15.jpg",
		}

		resp := MakeRequest(t, ts, "POST", "/api/products", productData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var product models.Product
		err := json.NewDecoder(resp.Body).Decode(&product)
		assert.NoError(t, err)

		assert.Equal(t, "iPhone 15", product.Name)
		assert.Equal(t, "Latest iPhone model", product.Description)
		assert.Equal(t, 999.99, product.Price)
		assert.Equal(t, category.ID, product.CategoryID)
		assert.Equal(t, 10, product.Stock)
	})

	t.Run("GetAllProducts", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)
		CreateTestProduct(t, ts, category.ID)

		resp := MakeRequest(t, ts, "GET", "/api/products", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var products []models.Product
		err := json.NewDecoder(resp.Body).Decode(&products)
		assert.NoError(t, err)

		assert.Greater(t, len(products), 0)
	})

	t.Run("GetProduct", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)

		resp := MakeRequest(t, ts, "GET", "/api/products/"+product.ID.String(), nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedProduct models.Product
		err := json.NewDecoder(resp.Body).Decode(&retrievedProduct)
		assert.NoError(t, err)

		assert.Equal(t, product.ID, retrievedProduct.ID)
		assert.Equal(t, product.Name, retrievedProduct.Name)
		assert.Equal(t, product.Price, retrievedProduct.Price)
	})

	t.Run("GetProductsByCategory", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)
		CreateTestProduct(t, ts, category.ID)

		resp := MakeRequest(t, ts, "GET", "/api/products/category/"+category.ID.String(), nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var products []models.Product
		err := json.NewDecoder(resp.Body).Decode(&products)
		assert.NoError(t, err)

		assert.Greater(t, len(products), 0)
		for _, product := range products {
			assert.Equal(t, category.ID, product.CategoryID)
		}
	})

	t.Run("GetAveragePriceByCategory", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)
		CreateTestProduct(t, ts, category.ID)

		resp := MakeRequest(t, ts, "GET", "/api/products/category/"+category.ID.String()+"/average-price", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var categoryPrice models.CategoryPrice
		err := json.NewDecoder(resp.Body).Decode(&categoryPrice)
		assert.NoError(t, err)

		assert.Equal(t, category.ID, categoryPrice.CategoryID)
		assert.Equal(t, category.Name, categoryPrice.CategoryName)
		assert.Greater(t, categoryPrice.AveragePrice, 0.0)
		assert.Greater(t, categoryPrice.ProductCount, 0)
	})
}

// TestOrderRoutes tests all order-related endpoints
func TestOrderRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateOrder", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)

		orderData := map[string]interface{}{
			"customer_id": customer.ID.String(),
			"items": []map[string]interface{}{
				{
					"product_id": product.ID.String(),
					"quantity":   2,
				},
			},
		}

		resp := MakeRequest(t, ts, "POST", "/api/orders", orderData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var order models.Order
		err := json.NewDecoder(resp.Body).Decode(&order)
		assert.NoError(t, err)

		assert.Equal(t, customer.ID, order.CustomerID)
		assert.Equal(t, "pending", order.Status)
		assert.Greater(t, order.Total, 0.0)
		assert.Equal(t, 1, len(order.Items))
	})

	t.Run("CreateOrderInvalidCustomer", func(t *testing.T) {
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)

		orderData := map[string]interface{}{
			"customer_id": uuid.New().String(),
			"items": []map[string]interface{}{
				{
					"product_id": product.ID.String(),
					"quantity":   2,
				},
			},
		}

		resp := MakeRequest(t, ts, "POST", "/api/orders", orderData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("CreateOrderInvalidProduct", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)

		orderData := map[string]interface{}{
			"customer_id": customer.ID.String(),
			"items": []map[string]interface{}{
				{
					"product_id": uuid.New().String(),
					"quantity":   2,
				},
			},
		}

		resp := MakeRequest(t, ts, "POST", "/api/orders", orderData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("GetOrder", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)
		order := CreateTestOrder(t, ts, customer.ID, product.ID)

		resp := MakeRequest(t, ts, "GET", "/api/orders/"+order.ID.String(), nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedOrder models.Order
		err := json.NewDecoder(resp.Body).Decode(&retrievedOrder)
		assert.NoError(t, err)

		assert.Equal(t, order.ID, retrievedOrder.ID)
		assert.Equal(t, order.CustomerID, retrievedOrder.CustomerID)
		assert.Equal(t, order.Status, retrievedOrder.Status)
		assert.Equal(t, order.Total, retrievedOrder.Total)
	})

	t.Run("GetOrdersByCustomer", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)
		CreateTestOrder(t, ts, customer.ID, product.ID)

		resp := MakeRequest(t, ts, "GET", "/api/customers/"+customer.ID.String()+"/orders", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var orders []models.Order
		err := json.NewDecoder(resp.Body).Decode(&orders)
		assert.NoError(t, err)

		assert.Greater(t, len(orders), 0)
		for _, order := range orders {
			assert.Equal(t, customer.ID, order.CustomerID)
		}
	})

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)
		order := CreateTestOrder(t, ts, customer.ID, product.ID)

		statusData := map[string]interface{}{
			"status": "processing",
		}

		resp := MakeRequest(t, ts, "PUT", "/api/orders/"+order.ID.String()+"/status", statusData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedOrder models.Order
		err := json.NewDecoder(resp.Body).Decode(&updatedOrder)
		assert.NoError(t, err)

		assert.Equal(t, "processing", updatedOrder.Status)
	})

	t.Run("UpdateOrderStatusInvalid", func(t *testing.T) {
		customer, _ := CreateTestCustomer(t, ts)
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)
		order := CreateTestOrder(t, ts, customer.ID, product.ID)

		statusData := map[string]interface{}{
			"status": "invalid_status",
		}

		resp := MakeRequest(t, ts, "PUT", "/api/orders/"+order.ID.String()+"/status", statusData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	resp := MakeRequest(t, ts, "GET", "/health", nil, nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResponse map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&healthResponse)
	assert.NoError(t, err)

	assert.Equal(t, "ok", healthResponse["status"])
}
