"""
Application Settings

Pydantic-based settings management with environment variable support.
All configuration is loaded from environment variables or .env file.
"""

from functools import lru_cache
from typing import Literal

from pydantic import Field, PostgresDsn, RedisDsn, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """
    Application settings.
    
    All settings are loaded from environment variables.
    See .env.example for available options.
    """
    
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
        extra="ignore",  # Ignore extra fields from .env
    )
    
    # ==================== Application ====================
    APP_NAME: str = Field(default="py-core", description="Application name")
    APP_VERSION: str = Field(default="0.1.0", description="Application version")
    DEBUG: bool = Field(default=False, description="Debug mode")
    ENVIRONMENT: Literal["development", "staging", "production"] = Field(
        default="development", 
        description="Environment name"
    )
    
    # ==================== Server ====================
    HOST: str = Field(default="0.0.0.0", description="Server host")
    PORT: int = Field(default=8000, description="Server port", ge=1, le=65535)
    WORKERS: int = Field(default=1, description="Number of worker processes", ge=1)
    RELOAD: bool = Field(default=False, description="Auto-reload on code changes")
    
    # ==================== Database ====================
    POSTGRES_USER: str = Field(description="PostgreSQL username")
    POSTGRES_PASSWORD: str = Field(description="PostgreSQL password")
    POSTGRES_HOST: str = Field(default="localhost", description="PostgreSQL host")
    POSTGRES_PORT: int = Field(default=5432, description="PostgreSQL port", ge=1, le=65535)
    POSTGRES_DB: str = Field(description="PostgreSQL database name")
    
    # Database connection pool settings
    DB_POOL_SIZE: int = Field(default=5, description="Database connection pool size", ge=1)
    DB_MAX_OVERFLOW: int = Field(default=10, description="Max overflow connections", ge=0)
    DB_POOL_TIMEOUT: int = Field(default=30, description="Pool timeout in seconds", ge=1)
    DB_POOL_RECYCLE: int = Field(default=3600, description="Connection recycle time", ge=1)
    DB_ECHO: bool = Field(default=False, description="Echo SQL queries")
    
    @property
    def DATABASE_URL(self) -> str:
        """Construct async PostgreSQL connection URL."""
        return (
            f"postgresql+asyncpg://{self.POSTGRES_USER}:{self.POSTGRES_PASSWORD}"
            f"@{self.POSTGRES_HOST}:{self.POSTGRES_PORT}/{self.POSTGRES_DB}"
        )
    
    @property
    def DATABASE_URL_SYNC(self) -> str:
        """Construct sync PostgreSQL connection URL (for Alembic)."""
        return (
            f"postgresql://{self.POSTGRES_USER}:{self.POSTGRES_PASSWORD}"
            f"@{self.POSTGRES_HOST}:{self.POSTGRES_PORT}/{self.POSTGRES_DB}"
        )
    
    # ==================== Redis ====================
    REDIS_HOST: str = Field(default="localhost", description="Redis host")
    REDIS_PORT: int = Field(default=6379, description="Redis port", ge=1, le=65535)
    REDIS_DB: int = Field(default=0, description="Redis database number", ge=0)
    REDIS_PASSWORD: str | None = Field(default=None, description="Redis password")
    REDIS_SSL: bool = Field(default=False, description="Use SSL for Redis")
    
    @property
    def REDIS_URL(self) -> str:
        """Construct Redis connection URL."""
        auth = f":{self.REDIS_PASSWORD}@" if self.REDIS_PASSWORD else ""
        scheme = "rediss" if self.REDIS_SSL else "redis"
        return f"{scheme}://{auth}{self.REDIS_HOST}:{self.REDIS_PORT}/{self.REDIS_DB}"
    
    # ==================== Security ====================
    SECRET_KEY: str = Field(
        min_length=32,
        description="Secret key for JWT tokens (min 32 characters)"
    )
    ALGORITHM: str = Field(default="HS256", description="JWT algorithm")
    ACCESS_TOKEN_EXPIRE_MINUTES: int = Field(
        default=30, 
        description="Access token expiration in minutes",
        ge=1
    )
    REFRESH_TOKEN_EXPIRE_DAYS: int = Field(
        default=7,
        description="Refresh token expiration in days",
        ge=1
    )
    
    # ==================== CORS ====================
    CORS_ORIGINS: list[str] = Field(
        default=["http://localhost:3000"],
        description="Allowed CORS origins"
    )
    CORS_ALLOW_CREDENTIALS: bool = Field(
        default=True,
        description="Allow credentials in CORS"
    )
    CORS_ALLOW_METHODS: list[str] = Field(
        default=["*"],
        description="Allowed HTTP methods"
    )
    CORS_ALLOW_HEADERS: list[str] = Field(
        default=["*"],
        description="Allowed HTTP headers"
    )
    
    # ==================== Celery ====================
    CELERY_BROKER_URL: str = Field(
        default="redis://localhost:6379/0",
        description="Celery broker URL"
    )
    CELERY_RESULT_BACKEND: str = Field(
        default="redis://localhost:6379/0",
        description="Celery result backend URL"
    )
    
    # ==================== Logging ====================
    LOG_LEVEL: Literal["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"] = Field(
        default="INFO",
        description="Logging level"
    )
    LOG_FORMAT: Literal["json", "console"] = Field(
        default="console",
        description="Log output format"
    )
    
    # ==================== Rate Limiting ====================
    RATE_LIMIT_ENABLED: bool = Field(default=True, description="Enable rate limiting")
    RATE_LIMIT_PER_MINUTE: int = Field(
        default=60,
        description="Max requests per minute",
        ge=1
    )
    
    # ==================== Validation ====================
    
    @field_validator("ENVIRONMENT")
    @classmethod
    def validate_environment(cls, v: str) -> str:
        """Ensure environment is lowercase."""
        return v.lower()
    
    @field_validator("SECRET_KEY")
    @classmethod
    def validate_secret_key(cls, v: str) -> str:
        """Ensure secret key is strong enough."""
        if len(v) < 32:
            raise ValueError("SECRET_KEY must be at least 32 characters long")
        if v == "your-secret-key-change-this-in-production-min-32-chars":
            if cls.model_fields.get("ENVIRONMENT", "development") == "production":
                raise ValueError("Must change SECRET_KEY in production!")
        return v
    
    # ==================== Helper Properties ====================
    
    @property
    def is_development(self) -> bool:
        """Check if running in development mode."""
        return self.ENVIRONMENT == "development"
    
    @property
    def is_production(self) -> bool:
        """Check if running in production mode."""
        return self.ENVIRONMENT == "production"
    
    @property
    def is_staging(self) -> bool:
        """Check if running in staging mode."""
        return self.ENVIRONMENT == "staging"


@lru_cache
def get_settings() -> Settings:
    """
    Get cached settings instance.
    
    This function is cached so that settings are only loaded once.
    Use this instead of creating Settings() directly.
    
    Returns:
        Settings instance
    """
    return Settings()


# Convenience exports
settings = get_settings()

__all__ = ["Settings", "get_settings", "settings"]