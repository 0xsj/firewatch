from enum import Enum 
from typing import Generic, TypeVar
from pydantic import BaseModel, Field

T = TypeVar("T")

class SortOrder(str, Enum):

    ASC = "asc"
    DESC = "desc"

class PaginationParams(BaseModel):

    page: int = Field(default=1, ge=1, description="Page number (1-indexed)")
    page_size: int = Field(default=20, ge=1, le=100, description="Items per page")
    sort_by: str | None = Field(default=None, description="Field to sort by")
    sort_order: SortOrder = Field(default=SortOrder.DESC, description="Sort direction")

    @property
    def skip(self) -> int:
        return (self.page - 1) * self.page_size

    @property
    def limit(self) -> int:
        return self.page_size

# TODO: pagination result 
# TODO: cursorpagination params
# TODO: cursor pagination result