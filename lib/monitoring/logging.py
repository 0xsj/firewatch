"""
Structured logging configuration for the application.
Provides JSON logging, correlation IDs, and security-focused log filtering.
"""
import logging
import logging.config
import sys
import uuid
from typing import Any, Dict, Optional
from contextvars import ContextVar
from datetime import datetime

import structlog
from structlog.types import Processor

# Context variable for request correlation ID
correlation_id: ContextVar[Optional[str]] = ContextVar('correlation_id', default=None)

# =============================================================================
# Log Filtering and Security
# =============================================================================

class PIIFilter(logging.Filter):
    """Filter to remove PII from logs."""
    
    SENSITIVE_FIELDS = {
        'password', 'token', 'secret', 'key', 'authorization',
        'email', 'phone', 'ssn', 'credit_card', 'api_key'
    }
    
    def filter(self, record: logging.LogRecord) -> bool:
        """Filter sensitive information from log records."""
        if hasattr(record, 'msg') and isinstance(record.msg, dict):
            record.msg = self._sanitize_dict(record.msg)
        return True
    
    def _sanitize_dict(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Remove sensitive fields from dictionary."""
        sanitized = {}
        for key, value in data.items():
            if any(sensitive in key.lower() for sensitive in self.SENSITIVE_FIELDS):
                sanitized[key] = "[REDACTED]"
            elif isinstance(value, dict):
                sanitized[key] = self._sanitize_dict(value)
            else:
                sanitized[key] = value
        return sanitized

# =============================================================================
# Custom Processors
# =============================================================================

def add_correlation_id(logger: Any, method_name: str, event_dict: Dict[str, Any]) -> Dict[str, Any]:
    """Add correlation ID to log entries."""
    corr_id = correlation_id.get()
    if corr_id:
        event_dict["correlation_id"] = corr_id
    return event_dict

def add_timestamp(logger: Any, method_name: str, event_dict: Dict[str, Any]) -> Dict[str, Any]:
    """Add ISO timestamp to log entries."""
    event_dict["timestamp"] = datetime.utcnow().isoformat() + "Z"
    return event_dict

def add_level(logger: Any, method_name: str, event_dict: Dict[str, Any]) -> Dict[str, Any]:
    """Add log level to event dict."""
    event_dict["level"] = method_name
    return event_dict

# =============================================================================
# Logger Configuration
# =============================================================================

def configure_logging(environment: str = "development") -> None:
    """
    Configure structured logging for the application.
    
    Args:
        environment: The environment (development, production, testing)
    """
    processors: list[Processor] = [
        add_timestamp,
        add_level,
        add_correlation_id,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
    ]
    
    # Environment-specific configuration
    if environment == "development":
        # Pretty console output for development
        processors.extend([
            structlog.dev.ConsoleRenderer(colors=True)
        ])
        log_level = "DEBUG"
        
    elif environment == "testing":
        # Minimal logging for tests
        processors.extend([
            structlog.processors.JSONRenderer()
        ])
        log_level = "WARNING"
        
    else:  # production
        # JSON logging for production
        processors.extend([
            structlog.processors.JSONRenderer()
        ])
        log_level = "INFO"
    
    # Configure structlog
    structlog.configure(
        processors=processors,
        wrapper_class=structlog.stdlib.BoundLogger,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )
    
    # Configure standard library logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, log_level),
    )
    
    # Add PII filter to root logger
    root_logger = logging.getLogger()
    root_logger.addFilter(PIIFilter())

# =============================================================================
# Logger Factory
# =============================================================================

def get_logger(name: str = None) -> structlog.BoundLogger:
    """
    Get a configured logger instance.
    
    Args:
        name: Logger name (usually __name__)
        
    Returns:
        Configured structlog logger
    """
    return structlog.get_logger(name or __name__)

# =============================================================================
# Correlation ID Management
# =============================================================================

def set_correlation_id(corr_id: str = None) -> str:
    """
    Set correlation ID for the current context.
    
    Args:
        corr_id: Correlation ID (generates new UUID if None)
        
    Returns:
        The correlation ID that was set
    """
    if not corr_id:
        corr_id = str(uuid.uuid4())
    
    correlation_id.set(corr_id)
    return corr_id

def get_correlation_id() -> Optional[str]:
    """Get the current correlation ID."""
    return correlation_id.get()

def clear_correlation_id() -> None:
    """Clear the correlation ID from current context."""
    correlation_id.set(None)

# =============================================================================
# Request Logging Helpers
# =============================================================================

class RequestLogger:
    """Helper class for HTTP request/response logging."""
    
    def __init__(self, logger: structlog.BoundLogger):
        self.logger = logger
    
    def log_request(
        self, 
        method: str, 
        path: str, 
        client_ip: str = None,
        user_agent: str = None,
        **kwargs
    ) -> None:
        """Log incoming HTTP request."""
        self.logger.info(
            "HTTP request received",
            method=method,
            path=path,
            client_ip=client_ip or "[unknown]",
            user_agent=user_agent or "[unknown]",
            **kwargs
        )
    
    def log_response(
        self,
        status_code: int,
        duration_ms: float,
        **kwargs
    ) -> None:
        """Log HTTP response."""
        level = "error" if status_code >= 400 else "info"
        
        getattr(self.logger, level)(
            "HTTP response sent",
            status_code=status_code,
            duration_ms=round(duration_ms, 2),
            **kwargs
        )