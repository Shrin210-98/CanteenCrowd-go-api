
# CanteenCrowd-go-api

**A multi-tenant Management System API built with Go, featuring Role-Based Access Control (RBAC)**

## Tech Stack

- **Language**: Go
- **Database**: PostgreSQL 17
- **SQL Compiler to Go**: sqlc (v1.27.0)
- **Database Driver**: pgx/v5 (pgxpool)
- **Router**: Standard net/http
- **JWT**: golang-jwt/jwt/v5
- **UUID**: github.com/google/uuid
- **Validation**: github.com/go-playground/validator/v10
- **Migrations**: golang-migrate
- **Container**: Docker & Docker Compose

## Features

- **Multi-Tenant Architecture**: Shared database with tenant isolation
- **Role-Based Access Control (RBAC)**: Users → Roles → Permissions with per-user overrides
- **JWT Authentication**: Secure token-based authentication with tenant context
- **Employee Management**: Complete CRUD operations for employees, departments, and positions
- **User Management**: Staff and guest user creation with employee linking
- **Permission System**: Nested permission structure with role templates and custom overrides
- **PostgreSQL 17**: Robust database with JSONB support for flexible permissions

## Project Structure
```
ccms.com/api/
├── cmd/
│ └── api/
│ └── main.go # Application entry point
│
├── internal/
│ ├── constants/
│ │ ├── config.go  # Application constants (timeouts, user types, CORS)
│ │ ├── defaults.go # Default values (pagination, lockout settings)
│ │ └── permissions.go # Default permission templates and helpers
│ │
│ ├── database/
│ │ ├── sql/
│ │ │ ├── schemas/ # Database table definitions (001_init.sql, 002_users.sql, etc.)
│ │ │ └── queries/  # SQL queries (users.sql, employees.sql, etc.)
│ │ ├── models.go # Generated: Database models
│ │ ├── querier.go # Generated: Querier interface
│ │ └── db.go # Generated: Database connection
│ │
│ ├── handlers/
│ │ ├── auth.go # Authentication handlers (Login, Register)
│ │ ├── handler.go # Main Handler struct and NewHandler
│ │ ├── db_health.go # Database health check handler (DatabaseHealth)
│ │ ├── employees/ # Employee management handlers
│ │ └── users/     # Users management handlers
│ │
│ ├── server/
│ │ ├── server.go # HTTP server setup and database connection
│ │ ├── routes.go # Route registration
│ │ └── middleware.go # Middleware (CORS, Logger, Auth)
│ │
│ └── utils/
|
├── migrations/
│ ├── 000001_init_schema.up.sql # Initial schema migration
│ └── 000001_init_schema.down.sql # Rollback migration
│
├── nginx/
│ └── nginx.conf # Nginx reverse proxy configuration
│
├── .github/
│ └── workflows/
│ └── ci.yml # GitHub Actions CI/CD workflow
│
├── docker-compose.yml # Docker Compose for production
├── docker-compose.prod.yml # Docker Compose for development
├── Dockerfile # Production Dockerfile
├── Dockerfile.dev # Development Dockerfile
├── Makefile # Build and development commands
├── sqlc.yaml # sqlc configuration
├── .env # Environment variables (not committed)
├── .env.example # Example environment variables
├── go.mod # Go module definition
├── go.sum # Go module checksums
└── README.md # Project documentation
```

## Multi-Tenant Architecture

### How It Works

1. **Registration**: Creates a new tenant with a `tenant_owner` user
2. **JWT Token**: Contains `tenantID` and `userType` claims
3. **Middleware**: Extracts tenant context from JWT
4. **Queries**: Always filter by `tenant_id` for data isolation

### Tenant Context Flow
JWT Token → Middleware → Context → Handler → SQL Query (WHERE tenant_id = $1)


## RBAC (Role-Based Access Control)

### Core Concepts
Users → Roles → Permissions

User (John) → Role (Staff) → Permissions (dashboard, employees.view)
User (Jane) → Role (Admin) → Permissions (all access)


### User Types

- **tenant_owner**: Main account holder (created during registration)
  - Full access to all features
  - Can create users, manage roles, and configure permissions
  
- **staff**: Internal staff (requires employeeId)
  - Linked to employee record
  - Permissions based on assigned role
  
- **guest**: External temporary access (no employeeId)
  - Limited access
  - No employee link required

### Role with Permission Override
Effective Permissions = Role Permissions + User Override

User has:
- role_id: Staff role (default permissions)
- permissions_override: Custom permission changes

Rules:
1. No override → Use role permissions
2. Has override → Role permissions + override changes
3. Override wins for conflicting permissions

## Makefile Commands
```
make run          # Run locally
make build        # Build binary
make watch        # Hot reload with Air
make dev          # Docker development
make migrate-up   # Run migrations
make migrate-down # Rollback migration
make sqlc         # Generate SQL code
make db-shell     # Access database shell
make db-reset     # Reset database
make test         # Run tests
```

## Setup Instructions

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 17
- Docker & Docker Compose (optional)
- Make (for Makefile commands)
- Air (optional, for hot reload)

1. Clone the Repository
```
git clone <repository-url>
cd ccms-api
```


2. Install Go Dependencies
```
go mod tidy
```

3. Configure Environment Variables
```
cp .env.example .env
Edit .env with your configuration:
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=canteen_admin
DB_PASSWORD=your_password
DB_DATABASE=canteen_db
DB_SSLMODE=disable
DB_SCHEMA=public
JWT_SECRET=your_jwt_secret
PORT=8080
APP_ENV=development
```

4. Start PostgreSQL Database
```
docker compose up -d postgres
OR if using local PostgreSQL
sudo systemctl start postgresql
```

5. Reset Database with Schema
```
make db-reset
```
This drops and recreates the schema, applies all table definitions

7. Generate SQL Code (sqlc)
```
make sqlc
```
Generates Go code from SQL queries

9. Build the Application
```
make build
```
Creates binary in ./bin/

11. Run the Application
```
make run
```
Server starts on http://localhost:8080

13. Verify Health
```
curl http://localhost:8080/api/v1/health
```

15. Test Registration (Happy Path)
```
curl -X POST http://localhost:8080/api/v1/register \
-H "Content-Type: application/json" \
-d '{
    "username": "owner.acme",
    "email": "owner@acme.com",
    "password": "AcmePass123!",
    "tenantName": "Acme Corp",
    "tenantSlug": "acme-corp",
    "tenantPlan": "free",
    "fullName": "Acme Owner"
}'
```
11. Test Login
```
curl -X POST http://localhost:8080/api/v1/login \
-H "Content-Type: application/json" \
-d '{
    "email": "owner@acme.com",
    "password": "AcmePass123!"
}'
```

## Development Commands

#### Hot reload development
make watch
Requires Air installed: go install github.com/air-verse/air@latest

#### Docker development
make dev

#### Access database shell
make db-shell

#### Reset database (development only)
make db-reset

#### Run tests
make test

#### Check for linting issues
make lint


## License

#### This project is licensed under the MIT License. See the LICENSE file for details.

