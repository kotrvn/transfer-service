import uuid
from pydantic import BaseModel, ConfigDict, EmailStr

class UserOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: uuid.UUID
    name: str
    email: EmailStr
    role: str

class UserCreate(BaseModel):
    name: str
    email: EmailStr
    password: str
    role: str = "passenger" # По умолчанию пассажир