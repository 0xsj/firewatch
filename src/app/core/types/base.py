from datetime import datetime
from typing import TypeVar, NewType
from uuid import UUID

ID = TypeVar("ID", int, str, UUID)

class TimestampMixin:
    created_at: datetime
    updated_at: datetime