# E-commerce API Postman Collection

This Postman collection contains all the API endpoints for the e-commerce application, including OpenID Connect (OIDC) authentication.

## Getting Started

1. Import the collection file `ecommerce-api.postman_collection.json` into Postman
2. Set up the environment variables:
   - `base_url`: The base URL of your API (default: `http://localhost:8181`)
   - `token`: JWT token received after login (OIDC or traditional)
   - `auth_code`: OIDC authorization code (for testing OIDC callback)
   - `state`: OIDC state parameter (for testing OIDC callback)
   - `customer_id`: UUID of a customer
   - `category_id`: UUID of a category
   - `product_id`: UUID of a product
   - `order_id`: UUID of an order

## Authentication

The application supports two authentication methods:

### 1. OpenID Connect (OIDC) - Recommended
- Industry-standard authentication protocol
- Supports multiple identity providers (Google, Azure AD, Auth0, Keycloak)
- More secure and scalable than traditional JWT

### 2. Traditional JWT Authentication
- Backward compatible with existing implementations
- Simple email-based authentication

### Authentication Requirements by Endpoint

#### Public Endpoints (No Authentication Required)
- `GET /api/auth/login` - Initiate OIDC login
- `GET /api/auth/callback` - Handle OIDC callback
- `POST /api/customers` - Create customer
- `POST /api/customers/login` - Traditional customer login
- `GET /api/products` - Get all products
- `GET /api/products/{id}` - Get product by ID
- `GET /api/products/category/{id}` - Get products by category
- `GET /api/products/category/{id}/average-price` - Get average price by category
- `GET /api/categories` - Get all categories
- `GET /api/categories/{id}` - Get category by ID
- `GET /api/categories/{id}/children` - Get category children
- `POST /api/categories` - Create category
- `POST /api/products` - Create product
- `POST /api/orders` - Create order
- `GET /api/orders/{id}` - Get order details
- `PUT /api/orders/{id}/status` - Update order status
- `GET /health` - Health check

#### Protected Endpoints (Require Authentication via OIDC or JWT)
- `GET /api/auth/userinfo` - Get current user information
- `POST /api/auth/logout` - Logout current user
- `GET /api/customers/{id}` - Get customer details
- `GET /api/customers/{id}/orders` - Get customer orders

### How to Use OIDC Authentication

1. **Initiate Login**: Call `GET /api/auth/login` to start the OIDC flow
2. **Complete Authentication**: Complete authentication with your identity provider
3. **Handle Callback**: The provider will redirect to `/api/auth/callback` with an authorization code
4. **Receive Token**: The callback will return a JWT token for API access
5. **Use Token**: Set the `token` environment variable and use it for protected endpoints

### How to Use Traditional JWT Authentication

1. Create a customer or login to get a JWT token
2. Copy the token from the response
3. Set the `token` environment variable in Postman
4. All subsequent requests to protected endpoints will automatically include the token

## Available Endpoints

### Authentication

- **OIDC Login** `GET /api/auth/login` *(Public - No authentication required)*
  - Initiates OIDC login flow
  - Redirects to identity provider
  - No request body required

- **OIDC Callback** `GET /api/auth/callback` *(Public - No authentication required)*
  - Handles OIDC callback from identity provider
  - Query parameters: `code` (authorization code), `state` (CSRF protection)
  - Returns JWT token and user information

- **OIDC Logout** `POST /api/auth/logout` *(Protected - Requires authentication)*
  - Logs out current user
  - Clears authentication cookies
  - Requires valid JWT token

- **Get User Info** `GET /api/auth/userinfo` *(Protected - Requires authentication)*
  - Returns current user information
  - Requires valid JWT token

- **Traditional Customer Login** `POST /api/customers/login` *(Public - No authentication required)*
  ```json
  {
    "email": "john@example.com"
  }
  ```

### Customers

- **Create Customer** `POST /api/customers` *(Public - No authentication required)*
  ```json
  {
    "email": "john@example.com",
    "name": "John Doe",
    "phone": "1234567890"
  }
  ```

- **Get Customer** `GET /api/customers/{id}` *(Protected - Requires authentication via OIDC or JWT)*
- **Get Customer Orders** `GET /api/customers/{id}/orders` *(Protected - Requires authentication via OIDC or JWT)*

### Categories

- **Create Category** `POST /api/categories` *(Public - No authentication required)*
  ```json
  {
    "name": "Electronics",
    "description": "Electronic devices and accessories"
  }
  ```

- **Create Subcategory** `POST /api/categories` *(Public - No authentication required)*
  ```json
  {
    "name": "Smartphones",
    "description": "Mobile phones and smartphones",
    "parent_id": "category_uuid"
  }
  ```

- **Get All Categories** `GET /api/categories` *(Public - No authentication required)*
- **Get Category** `GET /api/categories/{id}` *(Public - No authentication required)*
- **Get Category Children** `GET /api/categories/{id}/children` *(Public - No authentication required)*

### Products

- **Create Product** `POST /api/products` *(Public - No authentication required)*
  ```json
  {
    "name": "iPhone 15",
    "description": "Latest iPhone model",
    "price": 999.99,
    "category_id": "category_uuid",
    "stock": 10,
    "image_url": "https://example.com/iphone15.jpg"
  }
  ```

- **Get All Products** `GET /api/products` *(Public - No authentication required)*
- **Get Product** `GET /api/products/{id}` *(Public - No authentication required)*
- **Get Products by Category** `GET /api/products/category/{id}` *(Public - No authentication required)*
- **Get Average Price by Category** `GET /api/products/category/{id}/average-price` *(Public - No authentication required)*

### Orders

- **Create Order** `POST /api/orders` *(Public - No authentication required)*
  ```json
  {
    "customer_id": "customer_uuid",
    "items": [
      {
        "product_id": "product_uuid",
        "quantity": 2
      }
    ]
  }
  ```

- **Get Order** `GET /api/orders/{id}` *(Public - No authentication required)*
- **Update Order Status** `PUT /api/orders/{id}/status` *(Public - No authentication required)*
  ```json
  {
    "status": "processing"
  }
  ```

### Health Check

- **Health Check** `GET /health` *(Public - No authentication required)*

## Response Examples

### OIDC Callback Response
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "User Name",
    "picture": "https://example.com/avatar.jpg",
    "provider": "https://accounts.google.com"
  },
  "access_token": "jwt_token",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

### User Info Response
```json
{
  "user_id": "uuid",
  "email": "user@example.com"
}
```

### Customer Creation Response
```json
{
  "customer": {
    "id": "uuid",
    "email": "john@example.com",
    "name": "John Doe",
    "phone": "1234567890",
    "created_at": "2024-03-08T12:00:00Z",
    "updated_at": "2024-03-08T12:00:00Z"
  },
  "token": "jwt_token"
}
```

### Category Response
```json
{
  "id": "uuid",
  "name": "Electronics",
  "description": "Electronic devices and accessories",
  "parent_id": null,
  "level": 0,
  "path": "/Electronics",
  "created_at": "2024-03-08T12:00:00Z",
  "updated_at": "2024-03-08T12:00:00Z"
}
```

### Product Response
```json
{
  "id": "uuid",
  "name": "iPhone 15",
  "description": "Latest iPhone model",
  "price": 999.99,
  "category_id": "category_uuid",
  "stock": 10,
  "image_url": "https://example.com/iphone15.jpg",
  "created_at": "2024-03-08T12:00:00Z",
  "updated_at": "2024-03-08T12:00:00Z",
  "category": {
    "id": "uuid",
    "name": "Smartphones",
    "description": "Mobile phones and smartphones",
    "parent_id": "uuid",
    "level": 1,
    "path": "/Electronics/Smartphones"
  }
}
```

### Order Response
```json
{
  "id": "uuid",
  "customer_id": "customer_uuid",
  "status": "pending",
  "total": 1999.98,
  "created_at": "2024-03-08T12:00:00Z",
  "updated_at": "2024-03-08T12:00:00Z",
  "items": [
    {
      "id": "uuid",
      "order_id": "order_uuid",
      "product_id": "product_uuid",
      "quantity": 2,
      "price": 999.99,
      "product": {
        "id": "uuid",
        "name": "iPhone 15",
        "price": 999.99
      }
    }
  ]
}
```

### Average Price Response
```json
{
  "category_id": "uuid",
  "category_name": "Smartphones",
  "average_price": 899.99,
  "product_count": 5
}
```

## Error Responses

All error responses follow this format:
```json
{
  "error": "Error message describing what went wrong"
}
```

Common HTTP status codes:
- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing or invalid authentication
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Testing Flow

### OIDC Authentication Flow
1. Call `GET /api/auth/login` to initiate OIDC login
2. Complete authentication with your identity provider
3. Handle the callback and extract the JWT token
4. Use the token for protected endpoints

### Traditional Authentication Flow
1. Create a customer and save the JWT token
2. Use the token for protected endpoints

### General Testing Flow
1. Create a root category (e.g., Electronics)
2. Create a subcategory (e.g., Smartphones)
3. Create a product in the subcategory
4. Create an order for the product
5. Update the order status
6. Check the customer's orders
7. Get the average price for the category

## Environment Setup

1. Create a new environment in Postman
2. Add these variables:
   ```
   base_url: http://localhost:8181
   token: <empty>
   auth_code: <empty>
   state: <empty>
   customer_id: <empty>
   category_id: <empty>
   product_id: <empty>
   order_id: <empty>
   ```
3. After creating resources, update the corresponding variables with the returned UUIDs

## OIDC Configuration

To use OIDC authentication, you need to configure your identity provider:

1. **Google OAuth 2.0**: Set up OAuth 2.0 credentials in Google Cloud Console
2. **Azure AD**: Register application in Azure Active Directory
3. **Auth0**: Create application in Auth0 dashboard
4. **Keycloak**: Configure realm and client in Keycloak

See `examples/oidc-setup.md` for detailed setup instructions.

## Notes

- All timestamps are in ISO 8601 format
- All IDs are UUIDs
- The API supports both OIDC and JWT tokens for authentication
- Category paths show the full hierarchy (e.g., "/Electronics/Smartphones")
- Order status can be: "pending", "processing", "shipped", "delivered", "cancelled"
- OIDC provides enhanced security with CSRF protection and secure cookies
- Traditional JWT authentication remains available for backward compatibility
