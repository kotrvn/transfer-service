from typing import Annotated
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from app.core.db import get_session
from app.models.trip import Trip
from app.schemas.trip import TripCreate, TripOut
from app.api.deps import get_current_user # <-- Импортируем охранника
from app.models.user import User

router = APIRouter(prefix="/trips", tags=["trips"])
SessionDep = Annotated[AsyncSession, Depends(get_session)]

@router.get("", response_model=list[TripOut])
async def list_trips(session: SessionDep):
    result = await session.execute(select(Trip))
    return result.scalars().all()

@router.post("", response_model=TripOut, status_code=status.HTTP_201_CREATED)
async def create_trip(
    payload: TripCreate, 
    session: SessionDep,
    current_user: Annotated[User, Depends(get_current_user)] # <-- Только с токеном!
):
    # Проверка роли
    if current_user.role != "driver":
        raise HTTPException(status_code=403, detail="Только водители могут создавать поездки")

    new_trip = Trip(
        driver_id=current_user.id, # Берем ID из токена!
        from_city=payload.from_city,
        to_city=payload.to_city,
        departure_at=payload.departure_at,
        seats_total=payload.seats_total,
        seats_available=payload.seats_total,
        price_per_seat=payload.price_per_seat
    )
    session.add(new_trip)
    await session.commit()
    await session.refresh(new_trip)
    return new_trip