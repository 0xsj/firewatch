import logging
import logging.config
import sys
import uuid
from typing import Any, Dict, Optional
from contextvars import ContextVar
from datetime import datetime

import structlog
from structlog.types import Processor

correlation_id: ContextVar[Optional[str]] = ContextVar('correlation_id', default=None)