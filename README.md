# Hexagonal Architecture - Go

A production-ready implementation of Hexagonal Architecture (Ports & Adapters) in Go, featuring multi-tenant domains, event-driven communication, and comprehensive observability.

## 🏗️ Architecture Overview

This project demonstrates true hexagonal architecture with:

- **Framework-agnostic domains**: Business logic independent of HTTP frameworks, databases, or messaging systems
- **Ports & Adapters**: Clear separation between domain and infrastructure via interfaces
- **Domain-Driven Design**: Rich domain models with aggregates, value objects, and domain events
- **Dependency Inversion**: All dependencies point inward toward the domain
- **Event-Driven Architecture**: Loosely coupled domains communicating via events

## 📦 Project Structure
```
hexagonal-go/
├── cmd/
│   ├── api/                    # HTTP API application
│   │   ├── config/            # Application configuration
│   │   ├── main.go            # Entry point
│   │   ├── app.go             # App struct & dependencies
│   │   ├── router.go          # HTTP routing
│   │   └── wire.go            # Dependency injection
│   └── worker/                 # Background worker application
│       ├── config/            # Worker configuration
│       └── main.go            # Entry point
├── internal/                   # Private application code
│   ├── identity/              # Identity domain
│   │   ├── domain/            # Aggregates, entities, value objects
│   │   │   ├── user/          # User aggregate
│   │   │   ├── session/       # Session aggregate
│   │   │   ├── auth/          # Authentication value objects
│   │   │   └── oauth/         # OAuth entities
│   │   ├── application/       # Use cases
│   │   │   ├── command/       # Write operations
│   │   │   ├── query/         # Read operations
│   │   │   └── dto/           # Data transfer objects
│   │   ├── infrastructure/    # Adapters
│   │   │   └── repository/    # Database implementations
│   │   └── interface/         # Delivery mechanisms
│   │       └── http/v1/       # HTTP handlers
│   ├── tenant/                # Multi-tenancy domain
│   │   ├── domain/            # Tenant aggregate, plans, settings
│   │   ├── application/       # Commands & queries
│   │   ├── infrastructure/    # Repository implementations
│   │   └── interface/         # HTTP handlers
│   ├── audit/                 # Audit logging domain
│   │   ├── domain/            # Audit entry aggregate
│   │   ├── application/       # Event subscriber
│   │   └── infrastructure/    # Repository
│   ├── notifications/         # Notifications domain
│   │   ├── domain/            # Notification aggregate
│   │   ├── application/       # Event handlers
│   │   └── infrastructure/    # Repository
│   └── email/                 # Email template domain
│       ├── domain/            # Template aggregate
│       ├── application/       # CRUD commands & queries
│       ├── infrastructure/    # Repository
│       └── interface/         # HTTP handlers
├── pkg/                        # Shared infrastructure packages
│   ├── config/                # Configuration loading
│   ├── database/              # Database abstractions
│   │   └── postgres/          # PostgreSQL implementation
│   ├── cache/                 # Caching abstractions
│   ├── email/                 # Email sending & rendering
│   ├── errors/                # Structured error handling
│   ├── http/                  # HTTP utilities
│   │   ├── middleware/        # Auth, logging, metrics
│   │   └── response/          # Response helpers
│   ├── messaging/             # Event bus & pub/sub
│   ├── oauth/                 # OAuth providers
│   │   ├── google/            # Google OAuth
│   │   └── github/            # GitHub OAuth
│   ├── observability/         # Observability stack
│   │   ├── logger/            # Structured logging
│   │   ├── metrics/           # Prometheus metrics
│   │   └── tracing/           # OpenTelemetry tracing
│   ├── provider/              # Infrastructure providers (Wire)
│   ├── security/              # Security utilities
│   │   └── jwt/               # JWT service
│   ├── storage/               # File storage abstractions
│   └── types/                 # Common types (ID, Timestamp)
├── migrations/                 # Database migrations
├── deployments/
│   └── docker/                # Docker Compose for local dev
└── docs/
    └── swagger/               # OpenAPI documentation
```

## 🚀 Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Wire (for dependency injection): `go install github.com/google/wire/cmd/wire@latest`
- Swag (for API docs): `go install github.com/swaggo/swag/cmd/swag@latest`

### Installation
```bash
# Clone the repository
git clone https://github.com/0xsj/hexagonal-go.git
cd hexagonal-go

# Install dependencies
go mod download

# Start infrastructure
docker-compose -f deployments/docker/docker-compose.yml up -d

# Run database migrations
go run cmd/migrate/main.go up

# Generate Wire dependencies
go generate ./...

# Generate Swagger docs
swag init -g cmd/api/main.go -o docs/swagger

# Run the API
go run cmd/api/main.go
```

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=hexagonal
DB_PASSWORD=hexagonal_dev
DB_NAME=hexagonal_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key-min-32-chars
JWT_EXPIRES_IN=1h

# OAuth (optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Email (SMTP)
SMTP_HOST=localhost
SMTP_PORT=1025
EMAIL_FROM_ADDRESS=noreply@example.com
EMAIL_FROM_NAME=Hexagonal App

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Observability
LOG_LEVEL=debug
LOG_FORMAT=json
METRICS_ENABLED=true
TRACING_ENABLED=true
TRACING_ENDPOINT=http://localhost:4318
```

## 📚 Foundational Domains

### Identity

User management, authentication, and session handling.

**Features:**
- User registration with email verification
- Password-based authentication
- OAuth login (Google, GitHub)
- JWT access tokens with refresh token rotation
- Session management with optional Redis caching
- Role-based access (user, admin, moderator)
- Account suspension/reactivation

**Endpoints:**
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login with credentials
- `POST /api/v1/auth/logout` - Logout current session
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get current user
- `GET /api/v1/auth/oauth/{provider}` - OAuth login redirect
- `GET /api/v1/auth/oauth/{provider}/callback` - OAuth callback

### Tenant (Multi-Tenancy)

Organization/workspace management with subscription plans.

**Features:**
- Tenant lifecycle (create, suspend, reactivate, cancel, delete)
- Subscription plans (free, starter, pro, enterprise)
- Tenant-specific settings
- Owner assignment
- Billing integration support

**Status Transitions:**
```
trialing → active, suspended, cancelled, deleted
active   → suspended, cancelled, deleted
suspended → active, cancelled, deleted
cancelled → deleted
deleted  → (terminal)
```

**Endpoints:**
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants/{id}` - Get tenant
- `PUT /api/v1/tenants/{id}` - Update tenant
- `PUT /api/v1/tenants/{id}/settings` - Update settings
- `POST /api/v1/tenants/{id}/plan` - Change plan
- `POST /api/v1/tenants/{id}/suspend` - Suspend tenant
- `POST /api/v1/tenants/{id}/reactivate` - Reactivate tenant
- `DELETE /api/v1/tenants/{id}` - Delete tenant

### Audit

Automatic audit trail of all domain events.

**Features:**
- Subscribes to all domain events
- Records actor, action, resource, and context
- Supports filtering by tenant, user, event type, time range
- Stores event payload and metadata as JSONB

### Notifications

Event-driven notification delivery.

**Features:**
- Reacts to domain events (e.g., user registered)
- Multi-channel support (email, SMS, push)
- Template-based content

### Email Templates

Manageable email templates with localization.

**Features:**
- CRUD operations for templates
- Multi-locale support
- Template variables with validation
- Status lifecycle (draft, active, archived, deleted)
- Preview rendering with sample data

**Endpoints:**
- `POST /api/v1/email/templates` - Create template
- `GET /api/v1/email/templates` - List templates
- `GET /api/v1/email/templates/{id}` - Get template
- `GET /api/v1/email/templates/by-slug` - Get by slug
- `PUT /api/v1/email/templates/{id}` - Update template
- `POST /api/v1/email/templates/{id}/activate` - Activate
- `POST /api/v1/email/templates/{id}/archive` - Archive
- `DELETE /api/v1/email/templates/{id}` - Delete
- `POST /api/v1/email/templates/{id}/preview` - Preview

## 🎯 Core Principles

### 1. Dependency Rule

Dependencies only point inward. Domain layer has zero external dependencies.
```
Interface → Application → Domain
    ↓           ↓
Infrastructure (implements domain ports)
```

### 2. Ports & Adapters

**Ports** (interfaces in domain):
```go
// domain/user/repository.go
type Repository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id types.ID) (*User, error)
    FindByEmail(ctx context.Context, tenantID, email string) (*User, error)
}
```

**Adapters** (implementations in infrastructure):
```go
// infrastructure/repository/postgres_user_repository.go
type PostgresUserRepository struct {
    db database.DB
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *User) error {
    // PostgreSQL-specific implementation
}
```

### 3. Domain Events

Domains communicate via events, not direct calls:
```go
// Domain emits event
user.Register() // internally calls user.addEvent(NewUserRegistered(...))

// Application publishes after persistence
for _, event := range user.Events() {
    publisher.Publish(ctx, event)
}
user.ClearEvents()

// Other domains subscribe
type UserEventSubscriber struct { /* ... */ }
func (s *UserEventSubscriber) Handle(ctx context.Context, event messaging.Event) error {
    // React to user.registered event
}
```

### 4. CQRS Pattern

Commands (writes) and Queries (reads) are separated:
```go
// Commands mutate state
type RegisterUserCommand struct { /* dependencies */ }
func (c *RegisterUserCommand) Handle(ctx context.Context, req RegisterRequest) (*UserDTO, error)

// Queries read state
type GetUserQuery struct { /* dependencies */ }
func (q *GetUserQuery) Handle(ctx context.Context, userID types.ID) (*UserDTO, error)
```

### 5. Structured Errors

Rich error types with HTTP/gRPC mapping:
```go
err := errors.NotFound("UserRepository.FindByID", "user")
// Automatically maps to HTTP 404

err := errors.Validation("User.Create", "email", "invalid format")
// Automatically maps to HTTP 400
```

## 🔧 Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| HTTP Router | Chi |
| Database | PostgreSQL |
| Caching | Redis |
| DI Framework | Wire |
| Metrics | Prometheus |
| Tracing | OpenTelemetry + Jaeger |
| Logging | Structured (JSON/Console) |
| API Docs | Swagger/OpenAPI |
| Auth | JWT + OAuth2 |
| Email | SMTP + Mailpit (dev) |

## 🧪 Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific domain tests
go test ./internal/identity/...

# Run integration tests (requires Docker)
go test -tags=integration ./...
```

## 📖 API Documentation

When running locally, Swagger UI is available at:
- http://localhost:8080/swagger/index.html

## 🐳 Local Development
```bash
# Start all infrastructure
docker-compose -f deployments/docker/docker-compose.yml up -d

# Services available:
# - PostgreSQL: localhost:5432
# - Redis: localhost:6379
# - Mailpit: localhost:1025 (SMTP), localhost:8025 (UI)
# - Jaeger: localhost:16686 (UI)
# - Prometheus: localhost:9090

# Stop infrastructure
docker-compose -f deployments/docker/docker-compose.yml down
```

## 🔄 Adding a New Domain

1. Create domain structure:
```
   internal/newdomain/
   ├── domain/           # Aggregates, entities, value objects, events
   ├── application/
   │   ├── command/      # Write use cases
   │   ├── query/        # Read use cases
   │   └── dto/          # Data transfer objects
   ├── infrastructure/
   │   └── repository/   # Database implementations
   └── interface/
       └── http/v1/      # HTTP handlers
```

2. Define domain interfaces (ports) in `domain/`
3. Implement use cases in `application/`
4. Create adapters in `infrastructure/`
5. Add HTTP handlers in `interface/`
6. Create Wire provider set in `provider.go`
7. Add to `cmd/api/wire.go`
8. Add routes to `cmd/api/router.go`

## 📄 License

MIT