enterprise-python-monolith/
│
├── src/
│   └── app/
│       │
│       ├── main.py                      # FastAPI app factory
│       ├── asgi.py                      # ASGI entry point (production)
│       │
│       ├── core/                        # Shared foundation (framework code)
│       │   ├── config/
│       │   │   ├── __init__.py
│       │   │   ├── settings.py         # Pydantic settings
│       │   │   ├── database.py         # DB config
│       │   │   ├── cache.py            # Redis config
│       │   │   └── logging.py          # Logging config
│       │   │
│       │   ├── db/
│       │   │   ├── __init__.py
│       │   │   ├── base.py             # SQLAlchemy base
│       │   │   ├── session.py          # Session factory (async)
│       │   │   ├── mixins.py           # Model mixins (timestamps, soft delete)
│       │   │   └── types.py            # Custom column types
│       │   │
│       │   ├── cache/
│       │   │   ├── __init__.py
│       │   │   ├── backend.py          # Redis client
│       │   │   ├── decorators.py       # @cached, @invalidate_cache
│       │   │   └── keys.py             # Cache key patterns
│       │   │
│       │   ├── messaging/
│       │   │   ├── __init__.py
│       │   │   ├── events.py           # Event bus (in-memory + Redis pub/sub)
│       │   │   ├── publisher.py        # Event publisher
│       │   │   └── handlers.py         # Event handler registry
│       │   │
│       │   ├── security/
│       │   │   ├── __init__.py
│       │   │   ├── auth.py             # JWT handling
│       │   │   ├── password.py         # Password hashing (argon2)
│       │   │   ├── permissions.py      # RBAC/permissions
│       │   │   └── rate_limit.py       # Rate limiting
│       │   │
│       │   ├── observability/
│       │   │   ├── __init__.py
│       │   │   ├── logging.py          # Structured logging (structlog)
│       │   │   ├── metrics.py          # Prometheus metrics
│       │   │   ├── tracing.py          # OpenTelemetry tracing
│       │   │   └── health.py           # Health checks
│       │   │
│       │   ├── middleware/
│       │   │   ├── __init__.py
│       │   │   ├── correlation_id.py   # Request ID tracking
│       │   │   ├── logging.py          # Request/response logging
│       │   │   ├── error_handler.py    # Global error handling
│       │   │   ├── auth.py             # Auth middleware
│       │   │   └── metrics.py          # Metrics collection
│       │   │
│       │   ├── exceptions/
│       │   │   ├── __init__.py
│       │   │   ├── base.py             # Base exception classes
│       │   │   ├── http.py             # HTTP exceptions
│       │   │   └── handlers.py         # Exception handlers
│       │   │
│       │   └── utils/
│       │       ├── __init__.py
│       │       ├── datetime.py         # Timezone-aware datetime utils
│       │       ├── pagination.py       # Cursor/offset pagination
│       │       ├── validation.py       # Common validators
│       │       └── serialization.py    # JSON serialization
│       │
│       ├── modules/                     # Business domains (bounded contexts)
│       │   │
│       │   ├── users/                   # User management domain
│       │   │   ├── __init__.py
│       │   │   │
│       │   │   ├── models/             # SQLAlchemy models
│       │   │   │   ├── __init__.py
│       │   │   │   ├── user.py
│       │   │   │   └── profile.py
│       │   │   │
│       │   │   ├── schemas/            # Pydantic schemas (API contracts)
│       │   │   │   ├── __init__.py
│       │   │   │   ├── requests.py     # Request schemas
│       │   │   │   └── responses.py    # Response schemas
│       │   │   │
│       │   │   ├── repositories/       # Data access layer
│       │   │   │   ├── __init__.py
│       │   │   │   └── user.py
│       │   │   │
│       │   │   ├── services/           # Business logic
│       │   │   │   ├── __init__.py
│       │   │   │   ├── user.py
│       │   │   │   └── auth.py
│       │   │   │
│       │   │   ├── events/             # Domain events
│       │   │   │   ├── __init__.py
│       │   │   │   ├── user_created.py
│       │   │   │   └── user_updated.py
│       │   │   │
│       │   │   ├── handlers/           # Event handlers
│       │   │   │   ├── __init__.py
│       │   │   │   └── user_events.py
│       │   │   │
│       │   │   ├── dependencies.py     # FastAPI dependencies
│       │   │   ├── exceptions.py       # Domain exceptions
│       │   │   └── constants.py        # Domain constants
│       │   │
│       │   ├── orders/                  # Order management domain
│       │   │   ├── models/
│       │   │   │   ├── order.py
│       │   │   │   └── order_item.py
│       │   │   ├── schemas/
│       │   │   ├── repositories/
│       │   │   ├── services/
│       │   │   ├── events/
│       │   │   ├── handlers/
│       │   │   └── dependencies.py
│       │   │
│       │   ├── payments/                # Payment domain
│       │   │   ├── models/
│       │   │   ├── schemas/
│       │   │   ├── repositories/
│       │   │   ├── services/
│       │   │   │   ├── payment.py
│       │   │   │   └── stripe_adapter.py  # External service adapter
│       │   │   ├── events/
│       │   │   └── dependencies.py
│       │   │
│       │   ├── notifications/           # Notification domain
│       │   │   ├── models/
│       │   │   ├── schemas/
│       │   │   ├── repositories/
│       │   │   ├── services/
│       │   │   │   ├── notification.py
│       │   │   │   ├── email_provider.py
│       │   │   │   └── sms_provider.py
│       │   │   ├── templates/          # Email/SMS templates
│       │   │   │   ├── welcome.html
│       │   │   │   └── reset_password.html
│       │   │   └── dependencies.py
│       │   │
│       │   └── shared/                  # Shared between modules
│       │       ├── __init__.py
│       │       ├── models.py           # Shared models
│       │       └── schemas.py          # Shared schemas
│       │
│       ├── api/                         # API layer (presentation)
│       │   ├── __init__.py
│       │   ├── deps.py                 # Global API dependencies
│       │   │
│       │   ├── v1/                     # API version 1
│       │   │   ├── __init__.py
│       │   │   ├── router.py           # Aggregates all v1 routes
│       │   │   │
│       │   │   ├── users/
│       │   │   │   ├── __init__.py
│       │   │   │   └── endpoints.py    # User endpoints
│       │   │   │
│       │   │   ├── auth/
│       │   │   │   └── endpoints.py
│       │   │   │
│       │   │   ├── orders/
│       │   │   │   └── endpoints.py
│       │   │   │
│       │   │   ├── payments/
│       │   │   │   └── endpoints.py
│       │   │   │
│       │   │   └── health/
│       │   │       └── endpoints.py    # Health/readiness checks
│       │   │
│       │   └── v2/                     # API version 2 (future)
│       │       └── __init__.py
│       │
│       ├── workers/                     # Background tasks
│       │   ├── __init__.py
│       │   ├── celery_app.py           # Celery instance
│       │   ├── config.py               # Worker config
│       │   │
│       │   ├── tasks/                  # Task definitions
│       │   │   ├── __init__.py
│       │   │   ├── email.py            # Email tasks
│       │   │   ├── reports.py          # Report generation
│       │   │   ├── cleanup.py          # Cleanup tasks
│       │   │   └── webhooks.py         # Webhook delivery
│       │   │
│       │   └── schedules.py            # Periodic task schedules (Celery Beat)
│       │
│       ├── cli/                         # CLI commands (Click/Typer)
│       │   ├── __init__.py
│       │   ├── main.py                 # CLI entry point
│       │   │
│       │   └── commands/
│       │       ├── __init__.py
│       │       ├── db.py               # DB commands (migrate, seed)
│       │       ├── users.py            # User management
│       │       └── admin.py            # Admin tasks
│       │
│       └── integrations/               # External service clients
│           ├── __init__.py
│           ├── stripe/
│           │   ├── __init__.py
│           │   ├── client.py
│           │   └── webhooks.py
│           ├── sendgrid/
│           ├── aws/
│           │   ├── s3.py
│           │   └── sns.py
│           └── twilio/
│
├── alembic/                            # Database migrations
│   ├── versions/
│   │   ├── 001_create_users_table.py
│   │   ├── 002_create_orders_table.py
│   │   └── ...
│   ├── env.py                          # Migration environment
│   └── script.py.mako                  # Migration template
│
├── tests/                              # Tests (mirrors src structure)
│   ├── __init__.py
│   ├── conftest.py                     # Pytest fixtures
│   │
│   ├── unit/                           # Unit tests
│   │   ├── test_core/
│   │   │   ├── test_cache.py
│   │   │   └── test_security.py
│   │   │
│   │   └── test_modules/
│   │       ├── test_users/
│   │       │   ├── test_services.py
│   │       │   └── test_repositories.py
│   │       └── test_orders/
│   │
│   ├── integration/                    # Integration tests
│   │   ├── test_api/
│   │   │   ├── test_users_api.py
│   │   │   └── test_orders_api.py
│   │   │
│   │   └── test_workers/
│   │       └── test_email_tasks.py
│   │
│   ├── e2e/                           # End-to-end tests
│   │   └── test_user_flow.py
│   │
│   └── fixtures/                      # Test data
│       ├── users.py
│       ├── orders.py
│       └── factories.py               # Factory boy factories
│
├── scripts/                           # Utility scripts
│   ├── dev.sh                        # Start dev environment
│   ├── test.sh                       # Run tests
│   ├── lint.sh                       # Linting
│   ├── format.sh                     # Auto-format (black/ruff)
│   ├── seed_db.py                    # Seed database
│   └── generate_migration.sh         # Generate Alembic migration
│
├── deployments/                       # Deployment configurations
│   │
│   ├── docker/
│   │   ├── Dockerfile                # Production image
│   │   ├── Dockerfile.dev            # Development image
│   │   ├── docker-compose.yml        # Local development
│   │   ├── docker-compose.prod.yml   # Production compose
│   │   └── .dockerignore
│   │
│   ├── kubernetes/                    # K8s manifests
│   │   ├── base/
│   │   │   ├── deployment.yaml
│   │   │   ├── service.yaml
│   │   │   ├── configmap.yaml
│   │   │   ├── secrets.yaml
│   │   │   ├── ingress.yaml
│   │   │   ├── hpa.yaml              # Horizontal pod autoscaling
│   │   │   └── pdb.yaml              # Pod disruption budget
│   │   │
│   │   └── overlays/                  # Kustomize overlays
│   │       ├── dev/
│   │       ├── staging/
│   │       └── production/
│   │
│   ├── terraform/                     # Infrastructure as Code
│   │   ├── modules/
│   │   │   ├── vpc/
│   │   │   ├── rds/
│   │   │   ├── elasticache/
│   │   │   └── eks/
│   │   │
│   │   └── environments/
│   │       ├── dev/
│   │       ├── staging/
│   │       └── production/
│   │
│   └── helm/                          # Helm charts (alternative to k8s)
│       └── enterprise-app/
│           ├── Chart.yaml
│           ├── values.yaml
│           ├── values-dev.yaml
│           ├── values-prod.yaml
│           └── templates/
│
├── docs/                              # Documentation
│   ├── architecture/
│   │   ├── README.md                 # Architecture overview
│   │   ├── adr/                      # Architecture Decision Records
│   │   │   ├── 001-monolith-first.md
│   │   │   ├── 002-fastapi-choice.md
│   │   │   └── 003-event-driven.md
│   │   └── diagrams/
│   │       ├── system-context.png
│   │       └── module-dependencies.png
│   │
│   ├── api/
│   │   └── openapi.yaml              # Generated OpenAPI spec
│   │
│   ├── development/
│   │   ├── setup.md
│   │   ├── testing.md
│   │   ├── conventions.md
│   │   └── contributing.md
│   │
│   └── operations/
│       ├── deployment.md
│       ├── monitoring.md
│       ├── disaster-recovery.md
│       └── runbook.md
│
├── .github/                           # GitHub Actions
│   └── workflows/
│       ├── ci.yml                    # CI pipeline
│       ├── cd.yml                    # CD pipeline
│       ├── security-scan.yml         # Security scanning
│       └── dependency-update.yml     # Dependabot-like
│
├── .vscode/                           # VS Code settings
│   ├── settings.json
│   ├── launch.json                   # Debug configurations
│   └── extensions.json               # Recommended extensions
│
├── .env.example                       # Environment template
├── .env                               # Local env (gitignored)
├── .gitignore
├── .dockerignore
├── .pre-commit-config.yaml            # Pre-commit hooks
│
├── pyproject.toml                     # Project config (Poetry/Rye/uv)
├── poetry.lock / requirements.txt     # Dependency lock
├── alembic.ini                        # Alembic configuration
├── pytest.ini                         # Pytest configuration
├── ruff.toml                          # Ruff linter/formatter
├── mypy.ini                           # Type checking
├── coverage.ini                       # Coverage config
│
├── Makefile                           # Common commands
├── README.md
├── LICENSE
├── CHANGELOG.md
└── CONTRIBUTING.md