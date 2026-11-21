from typing import Annotated
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from app.core.db import get_session
from app.models.trip import Trip
from app.models.booking import Booking
from app.schemas.booking import BookingCreate, BookingOut
from app.api.deps import get_current_user
from app.models.user import User

router = APIRouter(prefix="/bookings", tags=["bookings"])
SessionDep = Annotated[AsyncSession, Depends(get_session)]

@router.post("", response_model=BookingOut, status_code=status.HTTP_201_CREATED)
async def create_booking(
    payload: BookingCreate, 
    session: SessionDep,
    current_user: Annotated[User, Depends(get_current_user)] 
):
    if current_user.role != "passenger":
        raise HTTPException(status_code=403, detail="Только пассажиры могут бронировать")

    # Ищем поездку и блокируем её (транзакция уже начата при проверке токена)
    stmt = select(Trip).where(Trip.id == payload.trip_id).with_for_update()
    result = await session.execute(stmt)
    trip = result.scalar_one_or_none()

    if not trip:
        raise HTTPException(status_code=404, detail="Поездка не найдена")
    
    if trip.seats_available < payload.seats:
        raise HTTPException(status_code=400, detail="Недостаточно свободных мест")

    # Уменьшаем количество мест
    trip.seats_available -= payload.seats
    
    # Создаем бронь
    new_booking = Booking(
        trip_id=payload.trip_id,
        passenger_id=current_user.id, 
        seats=payload.seats
    )
    session.add(new_booking)

    # Явно сохраняем всё в базу данных (вместо session.begin())
    await session.commit()
    
    # Обновляем объект, чтобы получить данные из базы
    await session.refresh(new_booking)
    
    return new_booking