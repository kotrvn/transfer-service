import uuid
from typing import Annotated
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import jwt, JWTError
from sqlalchemy.ext.asyncio import AsyncSession
from app.core.config import settings
from app.core.db import get_session
from app.models.user import User

# Эта строка говорит Swagger'у, куда стучаться за токеном
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/v1/users/login")

SessionDep = Annotated[AsyncSession, Depends(get_session)]

async def get_current_user(
    token: Annotated[str, Depends(oauth2_scheme)],
    session: SessionDep
) -> User:
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Не удалось проверить токен",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, settings.secret_key, algorithms=[settings.algorithm])
        user_id: str = payload.get("sub")
        if user_id is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception
    
    user = await session.get(User, uuid.UUID(user_id))
    if user is None:
        raise credentials_exception
    return user