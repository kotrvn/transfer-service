from datetime import datetime
import uuid
from sqlalchemy import String, Integer, ForeignKey, Numeric, DateTime
from sqlalchemy.orm import Mapped, mapped_column
from app.models.base import Base

class Trip(Base):
    __tablename__ = "trips"

    id: Mapped[uuid.UUID] = mapped_column(primary_key=True, default=uuid.uuid4)
    driver_id: Mapped[uuid.UUID] = mapped_column(ForeignKey("users.id"))
    
    from_city: Mapped[str] = mapped_column(String(100))
    to_city: Mapped[str] = mapped_column(String(100))
    departure_at: Mapped[datetime] = mapped_column(DateTime(timezone=True))
    
    seats_total: Mapped[int] = mapped_column(Integer)
    seats_available: Mapped[int] = mapped_column(Integer)
    
    # Цена за одно место (Numeric лучше для денег, чем Float)
    price_per_seat: Mapped[float] = mapped_column(Numeric(10, 2))
