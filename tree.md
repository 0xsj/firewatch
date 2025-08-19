py-core/
├── app/ # Application modules (feature-based)
│ ├── auth/
│ │ ├── **init**.py
│ │ ├── controller.py # FastAPI router
│ │ ├── service.py # Business logic
│ │ ├── repository.py # Data access
│ │ ├── models.py # Pydantic models & SQLAlchemy tables
│ │ ├── schemas.py # Request/Response schemas
│ │ ├── dependencies.py # FastAPI dependencies
│ │ ├── events.py # Domain events
│ │ └── tests/
│ │ ├── **init**.py
│ │ ├── test_controller.py
│ │ ├── test_service.py
│ │ └── test_repository.py
│ ├── users/
│ │ ├── **init**.py
│ │ ├── controller.py
│ │ ├── service.py
│ │ ├── repository.py
│ │ ├── models.py
│ │ ├── schemas.py
│ │ ├── dependencies.py
│ │ ├── events.py
│ │ └── tests/
│ ├── organizations/
│ └── **init**.py
├── lib/ # Reusable infrastructure
│ ├── cache/
│ │ ├── **init**.py
│ │ ├── service.py
│ │ ├── decorators.py # @cached decorator
│ │ ├── backends/
│ │ │ ├── **init**.py
│ │ │ ├── redis.py
│ │ │ └── memory.py
│ │ └── tests/
│ ├── queue/
│ │ ├── **init**.py
│ │ ├── service.py
│ │ ├── job.py # Base job class
│ │ ├── worker.py # Worker implementation
│ │ ├── backends/
│ │ │ ├── **init**.py
│ │ │ ├── celery_backend.py
│ │ │ ├── rq_backend.py
│ │ │ └── memory.py
│ │ └── tests/
│ ├── websocket/
│ │ ├── **init**.py
│ │ ├── manager.py # WebSocket connection manager
│ │ ├── auth.py # WebSocket authentication
│ │ ├── events.py # WebSocket event handlers
│ │ └── tests/
│ ├── database/
│ │ ├── **init**.py
│ │ ├── connection.py # Database connection/session
│ │ ├── base.py # Base repository/model classes
│ │ ├── migrations/
│ │ │ ├── **init**.py
│ │ │ ├── env.py # Alembic environment
│ │ │ ├── script.py.mako
│ │ │ └── versions/
│ │ ├── decorators.py # @transactional decorator
│ │ └── tests/
│ ├── events/
│ │ ├── **init**.py
│ │ ├── bus.py # Event bus implementation
│ │ ├── base.py # Base event/handler classes
│ │ ├── decorators.py # @event_handler decorator
│ │ └── tests/
│ ├── monitoring/
│ │ ├── **init**.py
│ │ ├── logging.py # Structured logging
│ │ ├── metrics.py # Prometheus metrics
│ │ ├── health/
│ │ │ ├── **init**.py
│ │ │ ├── service.py
│ │ │ ├── checks/
│ │ │ │ ├── **init**.py
│ │ │ │ ├── database.py
│ │ │ │ ├── redis.py
│ │ │ │ └── memory.py
│ │ │ └── tests/
│ │ └── tests/
│ ├── storage/
│ │ ├── **init**.py
│ │ ├── service.py
│ │ ├── backends/
│ │ │ ├── **init**.py
│ │ │ ├── local.py
│ │ │ ├── s3.py
│ │ │ └── gcs.py
│ │ └── tests/
│ ├── email/
│ │ ├── **init**.py
│ │ ├── service.py
│ │ ├── templates/
│ │ │ ├── **init**.py
│ │ │ └── base.py
│ │ ├── backends/
│ │ │ ├── **init**.py
│ │ │ ├── smtp.py
│ │ │ └── sendgrid.py
│ │ └── tests/
│ └── **init**.py
├── core/ # Application foundation
│ ├── **init**.py
│ ├── types.py # Common type definitions
│ ├── exceptions.py # Custom exceptions
│ ├── result.py # Result/Either monad
│ ├── config/
│ │ ├── **init**.py
│ │ ├── settings.py # Pydantic settings
│ │ └── environments/
│ │ ├── **init**.py
│ │ ├── development.py
│ │ ├── production.py
│ │ └── testing.py
│ ├── middleware/
│ │ ├── **init**.py
│ │ ├── auth.py
│ │ ├── cors.py
│ │ ├── logging.py
│ │ ├── rate_limit.py
│ │ └── error_handler.py
│ ├── security/
│ │ ├── **init**.py
│ │ ├── encryption.py
│ │ ├── hashing.py
│ │ └── jwt.py
│ ├── dependencies/ # FastAPI dependency injection
│ │ ├── **init**.py
│ │ ├── auth.py
│ │ ├── database.py
│ │ └── pagination.py
│ └── tests/
├── shared/ # Shared utilities
│ ├── **init**.py
│ ├── utils/
│ │ ├── **init**.py
│ │ ├── date.py
│ │ ├── string.py
│ │ ├── crypto.py
│ │ └── validation.py
│ ├── constants.py
│ ├── decorators.py # Common decorators
│ └── tests/
├── scripts/ # CLI scripts and tools
│ ├── **init**.py
│ ├── migrate.py # Database migrations
│ ├── seed.py # Data seeding
│ ├── worker.py # Queue worker runner
│ └── dev.py # Development utilities
├── tests/ # Global test configuration
│ ├── **init**.py
│ ├── conftest.py # pytest configuration
│ ├── fixtures/
│ │ ├── **init**.py
│ │ └── database.py
│ └── integration/
├── docs/ # Documentation
│ ├── api/ # API documentation (auto-generated)
│ └── deployment/
├── .env.example
├── .env.test
├── .gitignore
├── alembic.ini # Database migration config
├── docker-compose.yml
├── Dockerfile
├── pyproject.toml # Python project configuration
├── requirements.txt # Production dependencies
├── requirements-dev.txt # Development dependencies
├── main.py # Application entry point
└── README.md
