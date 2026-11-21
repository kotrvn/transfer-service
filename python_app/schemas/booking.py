import uuid
from pydantic import BaseModel, ConfigDict

class BookingCreate(BaseModel):
    # passenger_id: int  <-- ЭТОЙ СТРОКИ НЕ ДОЛЖНО БЫТЬ
    trip_id: uuid.UUID
    seats: int

class BookingOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: uuid.UUID
    trip_id: uuid.UUID
    passenger_id: uuid.UUID
    seats: int
    status: str