from fastapi import FastAPI
from app.api.users import router as users_router
from app.api.trips import router as trips_router
from app.api.bookings import router as bookings_router

app = FastAPI(title="Transfer Service")
app.include_router(users_router, prefix="/api/v1")
app.include_router(trips_router, prefix="/api/v1")
app.include_router(bookings_router, prefix="/api/v1")

@app.get("/health")
def health_check():
    return {"status": "ok"}