# Commerce Web Service

A modern e-commerce application built with Go, featuring a REST API, PostgreSQL database, Kubernetes deployment, and OpenID Connect authentication.

## Features

- **Customer Management**: User registration, authentication, and profile management
- **Product Catalog**: Hierarchical category system with unlimited depth
- **Order Management**: Complete order lifecycle with status tracking
- **SMS Notifications**: Order confirmations via Africa's Talking SMS gateway
- **Email Notifications**: Admin notifications for new orders
- **REST API**: Full-featured REST API with JWT authentication
- **OpenID Connect**: Industry-standard authentication with support for multiple identity providers
- **Multi-Provider Auth**: Support for Google, Azure AD, Auth0, Keycloak, and custom OIDC providers
- **Kubernetes Deployment**: Ready for production deployment on minikube

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │   PostgreSQL    │    │   SMS Gateway   │
│   (Go/Gin)      │◄──►│   Database      │    │   (Africa's     │
│   + OIDC Auth   │    │                 │    │   Talking)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   Kubernetes    │    │   Email Service │
│   (minikube)    │    │   (SMTP)        │
└─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐
│   OIDC          │
│   Providers     │
│   (Google,      │
│   Azure, etc.)  │
└─────────────────┘
```

## Authentication

The application supports two authentication methods:

### 1. OpenID Connect (OIDC)
- **Industry Standard**: Implements OpenID Connect 1.0 specification
- **Multiple Providers**: Google, Microsoft Azure AD, Auth0, Keycloak
- **Secure**: CSRF protection, state validation, secure cookies
- **Flexible**: Easy to add new identity providers

### 2. JWT Tokens
- **Backward Compatible**: Existing JWT authentication continues to work
- **Unified Middleware**: Single authentication layer for both OIDC and JWT
- **Seamless Integration**: OIDC users receive JWT tokens for API access

## Prerequisites

- Go 1.24 or later
- PostgreSQL 15 or later
- Docker
- minikube
- kubectl
- OIDC Provider (Google, Azure AD, Auth0, or Keycloak)

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd commerce-app
```

### 2. Set Up Environment Variables

Create a `.env` file in the root directory (refer to `.env.example`)

### 3. Configure OIDC Provider

See [OIDC Setup Guide](examples/oidc-setup.md) for detailed instructions on configuring:
- Google OAuth 2.0
- Microsoft Azure AD
- Auth0
- Keycloak
- Hydra By Ory

### 4. Run Locally

```bash
# Install dependencies
go mod tidy

# Run the application
go run cmd/server/main.go
```

The application will be available at `http://localhost:8181`

### 5. Test Authentication

```bash
# Test OIDC login
curl http://localhost:8181/api/auth/login

# Test protected endpoint (after authentication)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8181/api/auth/userinfo
```

### 6. Deploy to minikube

```bash
# Make sure minikube is running
minikube start

# Deploy the application
./deploy.sh
```

## API Documentation

### Authentication Endpoints

#### OIDC Authentication
- `GET /api/auth/login` - Initiate OIDC login
- `GET /api/auth/callback` - Handle OIDC callback
- `POST /api/auth/logout` - Logout user

#### Traditional JWT Authentication
- `POST /api/customers` - Register new customer
- `POST /api/customers/login` - Customer login

### Protected Endpoints

All protected endpoints require a valid JWT token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

## Security Features

- **CSRF Protection**: State parameter validation for OIDC flows
- **Secure Cookies**: HTTP-only, secure, SameSite attributes
- **Token Validation**: JWT and OIDC token verification
- **Role-Based Access**: Configurable RBAC system
- **Session Management**: Configurable timeouts and refresh
