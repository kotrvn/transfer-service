import asyncio
from app.core.db import engine
from app.models.base import Base

# Импортируем ВСЕ модели, чтобы SQLAlchemy знал о них
import app.models.user
import app.models.trip
import app.models.booking

async def create_tables():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

if __name__ == "__main__":
    asyncio.run(create_tables())
    print("✅ Таблицы успешно созданы (или обновлены) в базе данных!")