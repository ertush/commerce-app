package database

import (
	"time"

	"commerce-app/internal/models"

	"github.com/google/uuid"
)

// CustomerRepository handles customer database operations
type CustomerRepository struct{}

func (r *CustomerRepository) Create(customer *models.Customer) error {
	customer.ID = uuid.New()
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()

	query := `INSERT INTO customers (id, email, name, phone, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := DB.Exec(query, customer.ID, customer.Email, customer.Name, customer.Phone,
		customer.CreatedAt, customer.UpdatedAt)
	return err
}

func (r *CustomerRepository) GetByID(id uuid.UUID) (*models.Customer, error) {
	customer := &models.Customer{}
	query := `SELECT id, email, name, phone, created_at, updated_at FROM customers WHERE id = $1`

	err := DB.QueryRow(query, id).Scan(&customer.ID, &customer.Email, &customer.Name,
		&customer.Phone, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (r *CustomerRepository) GetByEmail(email string) (*models.Customer, error) {
	customer := &models.Customer{}
	query := `SELECT id, email, name, phone, created_at, updated_at FROM customers WHERE email = $1`

	err := DB.QueryRow(query, email).Scan(&customer.ID, &customer.Email, &customer.Name,
		&customer.Phone, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

// CategoryRepository handles category database operations
type CategoryRepository struct{}

func (r *CategoryRepository) Create(category *models.Category) error {
	category.ID = uuid.New()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	// Calculate level and path
	if category.ParentID != nil {
		parent, err := r.GetByID(*category.ParentID)
		if err != nil {
			return err
		}
		category.Level = parent.Level + 1
		category.Path = parent.Path + "/" + category.Name
	} else {
		category.Level = 0
		category.Path = "/" + category.Name
	}

	query := `INSERT INTO categories (id, name, description, parent_id, level, path, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := DB.Exec(query, category.ID, category.Name, category.Description,
		category.ParentID, category.Level, category.Path, category.CreatedAt, category.UpdatedAt)
	return err
}

func (r *CategoryRepository) GetByID(id uuid.UUID) (*models.Category, error) {
	category := &models.Category{}
	query := `SELECT id, name, description, parent_id, level, path, created_at, updated_at
			  FROM categories WHERE id = $1`

	err := DB.QueryRow(query, id).Scan(&category.ID, &category.Name, &category.Description,
		&category.ParentID, &category.Level, &category.Path, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	query := `SELECT id, name, description, parent_id, level, path, created_at, updated_at
			  FROM categories ORDER BY path`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description,
			&category.ParentID, &category.Level, &category.Path, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *CategoryRepository) GetChildren(parentID uuid.UUID) ([]models.Category, error) {
	query := `SELECT id, name, description, parent_id, level, path, created_at, updated_at
			  FROM categories WHERE parent_id = $1 ORDER BY name`

	rows, err := DB.Query(query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description,
			&category.ParentID, &category.Level, &category.Path, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// ProductRepository handles product database operations
type ProductRepository struct{}

func (r *ProductRepository) Create(product *models.Product) error {
	product.ID = uuid.New()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	query := `INSERT INTO products (id, name, description, price, category_id, stock, image_url, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := DB.Exec(query, product.ID, product.Name, product.Description, product.Price,
		product.CategoryID, product.Stock, product.ImageURL, product.CreatedAt, product.UpdatedAt)
	return err
}

func (r *ProductRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	query := `SELECT p.id, p.name, p.description, p.price, p.category_id, p.stock, p.image_url,
			  p.created_at, p.updated_at, c.id, c.name, c.description, c.parent_id, c.level, c.path,
			  c.created_at, c.updated_at
			  FROM products p
			  LEFT JOIN categories c ON p.category_id = c.id
			  WHERE p.id = $1`

	err := DB.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Description,
		&product.Price, &product.CategoryID, &product.Stock, &product.ImageURL,
		&product.CreatedAt, &product.UpdatedAt, &product.Category.ID, &product.Category.Name,
		&product.Category.Description, &product.Category.ParentID, &product.Category.Level,
		&product.Category.Path, &product.Category.CreatedAt, &product.Category.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	query := `SELECT p.id, p.name, p.description, p.price, p.category_id, p.stock, p.image_url,
			  p.created_at, p.updated_at, c.id, c.name, c.description, c.parent_id, c.level, c.path,
			  c.created_at, c.updated_at
			  FROM products p
			  LEFT JOIN categories c ON p.category_id = c.id
			  ORDER BY p.name`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description,
			&product.Price, &product.CategoryID, &product.Stock, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &product.Category.ID, &product.Category.Name,
			&product.Category.Description, &product.Category.ParentID, &product.Category.Level,
			&product.Category.Path, &product.Category.CreatedAt, &product.Category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *ProductRepository) GetByCategory(categoryID uuid.UUID) ([]models.Product, error) {
	query := `SELECT p.id, p.name, p.description, p.price, p.category_id, p.stock, p.image_url,
			  p.created_at, p.updated_at, c.id, c.name, c.description, c.parent_id, c.level, c.path,
			  c.created_at, c.updated_at
			  FROM products p
			  LEFT JOIN categories c ON p.category_id = c.id
			  WHERE p.category_id = $1
			  ORDER BY p.name`

	rows, err := DB.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description,
			&product.Price, &product.CategoryID, &product.Stock, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &product.Category.ID, &product.Category.Name,
			&product.Category.Description, &product.Category.ParentID, &product.Category.Level,
			&product.Category.Path, &product.Category.CreatedAt, &product.Category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// GetAveragePriceByCategory returns average price for a category
func (r *ProductRepository) GetAveragePriceByCategory(categoryID uuid.UUID) (*models.CategoryPrice, error) {
	query := `SELECT c.id, c.name, AVG(p.price) as average_price, COUNT(p.id) as product_count
			  FROM categories c
			  LEFT JOIN products p ON c.id = p.category_id
			  WHERE c.id = $1
			  GROUP BY c.id, c.name`

	var categoryPrice models.CategoryPrice
	err := DB.QueryRow(query, categoryID).Scan(&categoryPrice.CategoryID, &categoryPrice.CategoryName,
		&categoryPrice.AveragePrice, &categoryPrice.ProductCount)
	if err != nil {
		return nil, err
	}
	return &categoryPrice, nil
}

// OrderRepository handles order database operations
type OrderRepository struct{}

func (r *OrderRepository) Create(order *models.Order) error {
	order.ID = uuid.New()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	query := `INSERT INTO orders (id, customer_id, status, total, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(query, order.ID, order.CustomerID, order.Status, order.Total,
		order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert order items
	for i := range order.Items {
		order.Items[i].ID = uuid.New()
		order.Items[i].OrderID = order.ID

		itemQuery := `INSERT INTO order_items (id, order_id, product_id, quantity, price)
					  VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(itemQuery, order.Items[i].ID, order.Items[i].OrderID,
			order.Items[i].ProductID, order.Items[i].Quantity, order.Items[i].Price)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByID(id uuid.UUID) (*models.Order, error) {
	order := &models.Order{}
	query := `SELECT o.id, o.customer_id, o.status, o.total, o.created_at, o.updated_at,
			  c.id, c.email, c.name, c.phone, c.created_at, c.updated_at
			  FROM orders o
			  LEFT JOIN customers c ON o.customer_id = c.id
			  WHERE o.id = $1`

	err := DB.QueryRow(query, id).Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total,
		&order.CreatedAt, &order.UpdatedAt, &order.Customer.ID, &order.Customer.Email,
		&order.Customer.Name, &order.Customer.Phone, &order.Customer.CreatedAt, &order.Customer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get order items
	itemsQuery := `SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price,
				   p.id, p.name, p.description, p.price, p.category_id, p.stock, p.image_url,
				   p.created_at, p.updated_at
				   FROM order_items oi
				   LEFT JOIN products p ON oi.product_id = p.id
				   WHERE oi.order_id = $1`

	rows, err := DB.Query(itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price,
			&item.Product.ID, &item.Product.Name, &item.Product.Description, &item.Product.Price,
			&item.Product.CategoryID, &item.Product.Stock, &item.Product.ImageURL,
			&item.Product.CreatedAt, &item.Product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *OrderRepository) GetByCustomer(customerID uuid.UUID) ([]models.Order, error) {
	query := `SELECT o.id, o.customer_id, o.status, o.total, o.created_at, o.updated_at
			  FROM orders o
			  WHERE o.customer_id = $1
			  ORDER BY o.created_at DESC`

	rows, err := DB.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total,
			&order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
