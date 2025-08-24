from typing import TypeVar, Generic, Protocol, Any, Dict, List, Optional, Union
from datetime import datetime
from uuid import UUID

JsonDict = Dict[str, Any]
JsonList = List[JsonDict]
Headers = Dict[str, str]
QueryParams = Dict[str, str]

Email = str
TokenString = str

# Generics
T = TypeVar('T')
K = TypeVar('K')
V = TypeVar('V')
E = TypeVar('E', bound=Exception)

# universal protocols

class Timestamped(Protocol):
    """"""
    created_at: datetime
    updated_at: Optional[datetime]

class Cacheable(Protocol):
    """"""
    def cache_key(self) -> str:
        """Return a unique cache key for the object"""
        ...

class Serializable(Protocol):
    def to_dict(self) -> JsonDict:
        """"""
        ...

# utility
PaginationOffset = int
PaginationLimit = int
StatusCode = int
ResponseData = Union[JsonDict, JsonList, str, None]
AsyncResult = Any