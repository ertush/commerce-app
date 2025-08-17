<style>
.code-section {
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    background-color: #345;
    border: 1px solid #345;
    border-radius: 6px;
    padding: 10px;
    width: min-content;
    /*border: 2px solid #fff;*/
    white-space: collapse;
}
.btn {
    margin-end: 10px;
    align-self: end;
    width: min-content;
}

.btn:hover {
    background-color: #0056b3;
}

.btn-primary {

    background-color: #007bff;
    color: #fff;
    border: none;
    border-radius: 4px;
    padding: 5px 10px;
    cursor: pointer;
    hover: background-color: #0056b3;
}

.badge {
    color: #ffd;
    padding: 2px 2px;
    border-radius: 4px;
    border: 1px solid #ffd;"
}

.badge-success {
    color: #28a745;
    border: 1px solid #28a745;
}

.badge-danger {
    color: #dc3545;
    border: 1px solid #dc3545;
}

</style>

# E-Commerce API
![Commerce API Logo](https://img.freepik.com/premium-vector/ecommerce-logo-design_624194-152.jpg?w=200)

[<span style="color: #007bff; text-decoration: underline;">login</span>](https://ecommerce-app.eric-mutua.site/api/auth/login)
[<span style="color: #007bff; text-decoration: underline;">github</span>](https://github.com/ertush/commerce-app#readme)

This Postman collection contains all the API endpoints for the e-commerce application, including OpenID Connect (OIDC) authentication.

## Getting Started
1. Import the collection file `ecommerce-api.postman_collection.json` into Postman
2. Set up the environment variables:

  - <span class="badge">`base_url`</span> : The base URL of your API (default: `http://localhost:8181`)

  - <span class="badge">`token`</span> : JWT token received after login (OIDC)

  - <span class="badge">`auth_code`</span> : OIDC authorization code (for testing OIDC callback)

  - <span class="badge">`state`</span> : OIDC state parameter (for testing OIDC callback)

  - <span class="badge">`customer_id`</span> : UUID of a customer

  - <span class="badge">`category_id`</span> : UUID of a category

  - <span class="badge">`product_id`</span> : UUID of a product

  - <span class="badge">`order_id`</span> : UUID of an order

## Authentication

The app supports Open Connect ID oAuth2 authentication:

### 1. OpenID Connect (OIDC) - Recommended

- Industry-standard authentication protocol

- Supports multiple identity providers (Google, Azure AD, Ory, Auth0, Keycloak)


### Authentication Requirements by Endpoint

#### Public Endpoints (No Authentication Required)

- <span class="badge">`GET /api/auth/login`</span> - Initiate OIDC login

- <span class="badge">`GET /api/auth/callback`</span> - Handle OIDC callback

- <span class="badge">`POST /api/customers`</span> - Create customer

- <span class="badge">`POST /api/customers/login`</span> - Traditional customer login

- <span class="badge">`GET /api/products`</span> - Get all products

- <span class="badge">`GET /api/products/{id}`</span> - Get product by ID

- <span class="badge">`GET /api/products/category/{id}`</span> - Get products by category

- <span class="badge">`GET /api/products/category/{id}/average-price`</span> - Get average price by category

- <span class="badge">`GET /api/categories`</span> - Get all categories

- <span class="badge">`GET /api/categories/{id}`</span> - Get category by ID

- <span class="badge">`GET /api/categories/{id}/children`</span> - Get category children

- <span class="badge">`POST /api/categories`</span> - Create category

- <span class="badge">`POST /api/products`</span> - Create product

- <span class="badge">`POST /api/orders`</span> - Create order

- <span class="badge">`GET /api/orders/{id}`</span> - Get order details

- <span class="badge">`PUT /api/orders/{id}/status`</span> - Update order status

- <span class="badge">`GET /health`</span> - Health check

#### Protected Endpoints (Require Authentication via OIDC or JWT)

- <span class="badge">`GET /api/auth/userinfo`</span> - Get current user information

- <span class="badge">`POST /api/auth/logout`</span> - Logout current user

- <span class="badge">`GET /api/customers/{id}`</span> - Get customer details

- <span class="badge">`GET /api/customers/{id}/orders`</span> - Get customer orders

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

### Customers

- **Create Customer** `POST /api/customers` *(Public - No authentication required)*
  <div class="code-section" data-id="2">
  <button class="btn btn-primary" id="btn-1" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='2']").innerText.replace(/Copy/g,'')); console.log(document.querySelector("#btn-1").innerText)}>Copy</button>
  ```json
  {
    "email": "john@example.com",
    "name": "John Doe",
    "phone": "1234567890"
  }
  ```
  </div>

- **Get Customer** `GET /api/customers/{id}` *(Protected - Requires authentication via OIDC or JWT)*

- **Get Customer Orders** `GET /api/customers/{id}/orders` *(Protected - Requires authentication via OIDC or JWT)*

### Categories

- **Create Category** `POST /api/categories` *(Public - No authentication required)*
  <div class="code-section" data-id="3">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='3']").innerText.replace(/Copy/g,''))}>Copy</button>
  ```json
  {
    "name": "Electronics",
    "description": "Electronic devices and accessories"
  }
  ```
  </div>

- **Create Subcategory** `POST /api/categories` *(Public - No authentication required)*
  <div class="code-section" data-id="4">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='4']").innerText.replace(/Copy/g,''))}>Copy</button>
  ```json
  {
    "name": "Smartphones",
    "description": "Mobile phones and smartphones",
    "parent_id": "category_uuid"
  }
  ```
  </div>

- **Get All Categories** `GET /api/categories` *(Public - No authentication required)*
- **Get Category** `GET /api/categories/{id}` *(Public - No authentication required)*
- **Get Category Children** `GET /api/categories/{id}/children` *(Public - No authentication required)*

### Products

- **Create Product** `POST /api/products` *(Public - No authentication required)*
  <div class="code-section" data-id="5">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='5']").innerText.replace(/Copy/g,''))}>Copy</button>
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
  </div>

- **Get All Products** `GET /api/products` *(Public - No authentication required)*
- **Get Product** `GET /api/products/{id}` *(Public - No authentication required)*
- **Get Products by Category** `GET /api/products/category/{id}` *(Public - No authentication required)*
- **Get Average Price by Category** `GET /api/products/category/{id}/average-price` *(Public - No authentication required)*

### Orders

- **Create Order** `POST /api/orders` *(Public - No authentication required)*
  <div class="code-section" data-id="6">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='6']").innerText.replace(/Copy/g,''))}>Copy</button>
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
</div>

- **Get Order** `GET /api/orders/{id}` *(Public - No authentication required)*
- **Update Order Status** `PUT /api/orders/{id}/status` *(Public - No authentication required)*
  <div class="code-section" data-id="7">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='7']").innerText.replace(/Copy/g,''))}>Copy</button>
  ```json
  {
    "status": "processing"
  }
  ```
  </div>

### Health Check

- **Health Check** `GET /health` *(Public - No authentication required)*

## Response Examples


### OIDC Callback Response
  <div class="code-section" data-id="8">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='8']").innerText.replace(/Copy/g,''))}>Copy</button>
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
  </div>

### User Info Response
  <div class="code-section" data-id="9">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='9']").innerText.replace(/Copy/g,''))}>Copy</button>
  ```json
  {
  "user_id": "uuid",
  "email": "user@example.com"
  }
  ```
  </div>

### Customer Creation Response
  <div class="code-section" data-id="10">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='10']").innerText.replace(/Copy/g,''))}>Copy</button>
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
  </div>

### Category Response
  <div class="code-section" data-id="11">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='11']").innerText.replace(/Copy/g,''))}>Copy</button>
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
  </div>

### Product Response
  <div class="code-section" data-id="12">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='12']").innerText.replace(/Copy/g,''))}>Copy</button>
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
  </div>

### Order Response
  <div class="code-section" data-id="13">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='13']").innerText.replace(/Copy/g,''))}>Copy</button>
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
  </div>

### Average Price Response
  <div class="code-section" data-id="14">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='14']").innerText.replace(/Copy/g,''))}>Copy</button>
  ```json
  {
  "category_id": "uuid",
  "category_name": "Smartphones",
  "average_price": 899.99,
  "product_count": 5
  }
  ```
  </div>

## Error Responses

All error responses follow this format:
  <div class="code-section" data-id="15">
  ```json
  {
  "error": "Error message describing what went wrong"
  }
  ```
  </div>

Common HTTP status codes:

- <span class="badge badge-success">`200 OK`</span> : Request successful

- <span class="badge badge-success">`201 Created`</span> : Resource created successfully

- <span class="badge badge-danger">`400 Bad Request`</span> : Invalid request data

- <span class="badge badge-danger">`401 Unauthorized`</span> : Missing or invalid authentication

- <span class="badge badge-danger">`404 Not Found`</span> : Resource not found

- <span class="badge badge-danger">`500 Internal Server Error`</span> : Server error

## Testing Flow

### OIDC Authentication Flow

1. Call <span class="badge">`GET /api/auth/login`</span> to initiate OIDC login

2. Complete authentication with your identity provider

3. Handle the callback and extract the JWT token

4. Use the token for protected endpoints


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
  <div class="code-section" data-id="14">
  <button class="btn btn-primary" onclick={navigator.clipboard.writeText(document.querySelector("div[data-id='14']").innerText.replace(/Copy/g,''))}>Copy</button>
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
   </div>

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
