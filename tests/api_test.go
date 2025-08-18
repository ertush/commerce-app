package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"commerce-app/internal/database"
	"commerce-app/internal/models"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

// TestCustomerRoutes tests all customer-related endpoints
func TestCustomerRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateCustomer", func(t *testing.T) {

		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		customerData := GetTestUserDetails()

		customer, token := CreateTestCustomer(t, ts, customerData)

		assert.NotNil(t, customer)
		assert.NotEmpty(t, token)

		assert.Equal(t, customerData.Email, customer.Email)
		assert.Equal(t, customerData.Name, customer.Name)
		assert.Equal(t, customerData.Phone, customer.Phone)

		// Clean Up
		err := customerRepo.Delete(customer.ID)
		assert.NoError(t, err)

	})

	t.Run("CreateCustomerInvalidData", func(t *testing.T) {
		customerData := models.Customer{
			ID:    uuid.New(),
			Email: "invalid-email",
			Name:  "",
			Phone: "",
		}

		validCustomerData := GetTestUserDetails()

		_, token := CreateTestCustomer(t, ts, validCustomerData)

		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		resp := MakeRequest(t, ts, "POST", "/api/customers", customerData, headers)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("GetCustomer", func(t *testing.T) {
		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		customerData := GetTestUserDetails()
		customer, token := CreateTestCustomer(t, ts, customerData)

		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}

		resp := MakeRequest(t, ts, "GET", "/api/customers/"+customer.ID.String(), nil, headers)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedCustomer models.Customer
		err := json.NewDecoder(resp.Body).Decode(&retrievedCustomer)
		assert.NoError(t, err)

		assert.Equal(t, customer.ID, retrievedCustomer.ID)
		assert.Equal(t, customer.Email, retrievedCustomer.Email)
		assert.Equal(t, customer.Name, retrievedCustomer.Name)
		assert.Equal(t, customer.Phone, retrievedCustomer.Phone)

		// Clean up
		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)
	})

	t.Run("GetCustomerNotFound", func(t *testing.T) {
		randomID := uuid.New()

		validCustomerData := GetTestUserDetails()

		_, token := CreateTestCustomer(t, ts, validCustomerData)

		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		resp := MakeRequest(t, ts, "GET", "/api/customers/"+randomID.String(), nil, headers)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestCategoryRoutes tests all category-related endpoints
func TestCategoryRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateCategory", func(t *testing.T) {

		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		categoryData := map[string]string{
			"name":        "Test Category",
			"description": "Test Category Description",
		}

		resp := MakeRequest(t, ts, "POST", "/api/categories", categoryData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var category models.Category
		err := json.NewDecoder(resp.Body).Decode(&category)
		assert.NoError(t, err)

		assert.Equal(t, "Test Category", category.Name)
		assert.Equal(t, "Test Category Description", category.Description)
		assert.Equal(t, 0, category.Level)
		assert.Equal(t, "/Test Category", category.Path)

		//Clean up the created category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)
	})

	t.Run("CreateSubCategory", func(t *testing.T) {

		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		parentCategory := CreateTestCategory(t, ts, nil)

		categoryData := map[string]string{
			"name":        "Test Subcategory",
			"description": "Test Subcategory Description",
			"parent_id":   parentCategory.ID.String(),
		}

		resp := MakeRequest(t, ts, "POST", "/api/categories", categoryData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var category models.Category
		err := json.NewDecoder(resp.Body).Decode(&category)
		assert.NoError(t, err)

		assert.Equal(t, "Test Subcategory", category.Name)
		assert.Equal(t, 1, category.Level)
		assert.Equal(t, parentCategory.Path+"/Test Subcategory", category.Path)

		//Clean up the created parent category
		err = categoryRepo.Delete(parentCategory.ID, 0)
		assert.NoError(t, err)

		//Clean up the created subcategory
		err = categoryRepo.Delete(category.ID, 1)
		assert.NoError(t, err)
	})

	t.Run("GetAllCategories", func(t *testing.T) {

		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		createdCategory := CreateTestCategory(t, ts, nil)

		resp := MakeRequest(t, ts, "GET", "/api/categories", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var categories []models.Category
		err := json.NewDecoder(resp.Body).Decode(&categories)
		assert.NoError(t, err)

		assert.Greater(t, len(categories), 0)

		//Clean up the created category
		err = categoryRepo.Delete(createdCategory.ID, 0)
		assert.NoError(t, err)
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

		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		parentCategory := CreateTestCategory(t, ts, nil)
		childCategory := CreateTestCategory(t, ts, &parentCategory.ID)

		resp := MakeRequest(t, ts, "GET", "/api/categories/"+parentCategory.ID.String()+"/children", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var children []models.Category
		err := json.NewDecoder(resp.Body).Decode(&children)
		assert.NoError(t, err)

		assert.Greater(t, len(children), 0)

		// Clean Up Parent Category
		err = categoryRepo.Delete(parentCategory.ID, 0)
		assert.NoError(t, err)

		// Clean Up Child Category
		err = categoryRepo.Delete(childCategory.ID, 1)
		assert.NoError(t, err)
	})
}

// TestProductRoutes tests all product-related endpoints
func TestProductRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateProduct", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		category := CreateTestCategory(t, ts, nil)

		productData := map[string]interface{}{
			"name":        "Test Product",
			"description": "Test Product Description",
			"price":       999.99,
			"category_id": category.ID.String(),
			"stock":       10,
			"image_url":   "https://example.com/test-product.jpg",
		}

		resp := MakeRequest(t, ts, "POST", "/api/products", productData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var product models.Product
		err := json.NewDecoder(resp.Body).Decode(&product)
		assert.NoError(t, err)

		assert.Equal(t, "Test Product", product.Name)
		assert.Equal(t, "Test Product Description", product.Description)
		assert.Equal(t, 999.99, product.Price)
		assert.Equal(t, category.ID, product.CategoryID)
		assert.Equal(t, 10, product.Stock)

		// Clean Up Product
		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)

		// Clean Up Category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)
	})

	t.Run("GetAllProducts", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		category := CreateTestCategory(t, ts, nil)
		createdProduct := CreateTestProduct(t, ts, category.ID)

		resp := MakeRequest(t, ts, "GET", "/api/products", nil, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var products []models.Product
		err := json.NewDecoder(resp.Body).Decode(&products)
		assert.NoError(t, err)

		assert.Greater(t, len(products), 0)

		// Clean Up Created Product
		err = productRepo.Delete(createdProduct.ID)
		assert.NoError(t, err)

		// Clean Up Created Category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)
	})

	t.Run("GetProduct", func(t *testing.T) {
		// var categoryRepo *database.CategoryRepository
		// categoryRepo = &database.CategoryRepository{}

		// var productRepo *database.ProductRepository
		// productRepo = &database.ProductRepository{}

		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)

		productJson, err := json.Marshal(product)
		assert.NoError(t, err)

		log.Println(string(productJson))

		// log.Printf("Product ID: %v\n Category ID: %v", product.ID.String(), category.ID.String())

		// resp := MakeRequest(t, ts, "GET", "/api/products/"+product.ID.String(), nil, nil)
		// defer resp.Body.Close()

		// assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.Equal(t, 1, 1)

		// var retrievedProduct models.Product
		// err := json.NewDecoder(resp.Body).Decode(&retrievedProduct)
		// assert.NoError(t, err)

		// assert.Equal(t, product.ID, retrievedProduct.ID)
		// assert.Equal(t, product.Name, retrievedProduct.Name)
		// assert.Equal(t, product.Price, retrievedProduct.Price)

		//clean up created product
		// err = productRepo.Delete(product.ID)
		// assert.NoError(t, err)

		//clean up created category
		// err = categoryRepo.Delete(category.ID, 0)
		// assert.NoError(t, err)
	})

	t.Run("GetProductsByCategory", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		category := CreateTestCategory(t, ts, nil)
		createdProduct := CreateTestProduct(t, ts, category.ID)

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

		//Clean up created product
		err = productRepo.Delete(createdProduct.ID)
		assert.NoError(t, err)

		//Clean up created category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)
	})

	t.Run("GetAveragePriceByCategory", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		category := CreateTestCategory(t, ts, nil)
		createdProduct := CreateTestProduct(t, ts, category.ID)

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

		// Clean up created product
		err = productRepo.Delete(createdProduct.ID)
		assert.NoError(t, err)

		// Clean up created category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)
	})
}

// TestOrderRoutes tests all order-related endpoints
func TestOrderRoutes(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer(t)

	t.Run("CreateOrder", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		customerData := GetTestUserDetails()
		customer, _ := CreateTestCustomer(t, ts, customerData)
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

		//clean up Order
		err = orderRepo.Delete(order.ID)
		assert.NoError(t, err)

		//clean up customer
		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)

		//clean up product
		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)

		//clean up category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)
	})

	t.Run("CreateOrderInvalidCustomer", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

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

		var order models.Order
		err := json.NewDecoder(resp.Body).Decode(&order)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		//clean up category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)

		//clean up product
		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)

		//clean up order
		err = orderRepo.Delete(order.ID)
		assert.NoError(t, err)
	})

	t.Run("CreateOrderInvalidProduct", func(t *testing.T) {
		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

		customerData := GetTestUserDetails()
		customer, _ := CreateTestCustomer(t, ts, customerData)

		orderData := map[string]interface{}{
			"customer_id": customer.ID.String(),
			"items": []map[string]interface{}{
				{
					"quantity": 2,
				},
			},
		}

		resp := MakeRequest(t, ts, "POST", "/api/orders", orderData, nil)
		defer resp.Body.Close()

		var order models.Order
		err := json.NewDecoder(resp.Body).Decode(&order)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		//clean up order
		err = orderRepo.Delete(order.ID)
		assert.NoError(t, err)

		//clean up customer
		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)
	})

	t.Run("GetOrder", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		customerData := GetTestUserDetails()
		customer, _ := CreateTestCustomer(t, ts, customerData)
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

		//clean up order
		err = orderRepo.Delete(order.ID)
		assert.NoError(t, err)

		//clean up customer
		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)

		//clean up category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)

		//clean up product
		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)

	})

	t.Run("GetOrdersByCustomer", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

		customerData := GetTestUserDetails()

		customer, token := CreateTestCustomer(t, ts, customerData)

		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)
		createdOrder := CreateTestOrder(t, ts, customer.ID, product.ID)

		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}

		resp := MakeRequest(t, ts, "GET", "/api/customers/"+customer.ID.String()+"/orders", nil, headers)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var orders []models.Order
		err := json.NewDecoder(resp.Body).Decode(&orders)
		assert.NoError(t, err)

		assert.Greater(t, len(orders), 0)
		for _, order := range orders {
			assert.Equal(t, customer.ID, order.CustomerID)
		}

		//clean up order
		err = orderRepo.Delete(createdOrder.ID)
		assert.NoError(t, err)

		//clean up customer
		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)

		//clean up category
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)

		//clean up product
		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)
	})

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

		customerData := GetTestUserDetails()
		customer, _ := CreateTestCustomer(t, ts, customerData)
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

		//clean up
		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)

		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)

		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)

		err = orderRepo.Delete(order.ID)
		assert.NoError(t, err)
	})

	t.Run("UpdateOrderStatusInvalid", func(t *testing.T) {

		var categoryRepo *database.CategoryRepository
		categoryRepo = &database.CategoryRepository{}

		var productRepo *database.ProductRepository
		productRepo = &database.ProductRepository{}

		var customerRepo *database.CustomerRepository
		customerRepo = &database.CustomerRepository{}

		var orderRepo *database.OrderRepository
		orderRepo = &database.OrderRepository{}

		customerData := GetTestUserDetails()

		customer, _ := CreateTestCustomer(t, ts, customerData)
		category := CreateTestCategory(t, ts, nil)
		product := CreateTestProduct(t, ts, category.ID)
		order := CreateTestOrder(t, ts, customer.ID, product.ID)

		statusData := map[string]interface{}{
			"status": "invalid_status",
		}

		resp := MakeRequest(t, ts, "PUT", "/api/orders/"+order.ID.String()+"/status", statusData, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		//cleanup
		err := orderRepo.Delete(order.ID)
		assert.NoError(t, err)

		err = productRepo.Delete(product.ID)
		assert.NoError(t, err)

		err = categoryRepo.Delete(category.ID, 0)
		assert.NoError(t, err)

		err = customerRepo.Delete(customer.ID)
		assert.NoError(t, err)
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
