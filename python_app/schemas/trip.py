from datetime import datetime
import uuid
from pydantic import BaseModel, ConfigDict

class TripCreate(BaseModel):
    from_city: str
    to_city: str
    departure_at: datetime
    seats_total: int
    price_per_seat: float

class TripOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: uuid.UUID
    driver_id: uuid.UUID
    from_city: str
    to_city: str
    departure_at: datetime
    seats_total: int
    seats_available: int
    price_per_seat: float