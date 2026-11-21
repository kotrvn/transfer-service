package models

import (
	"time"
	"uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"` // "-" означает, что пароль НЕ попадет в JSON ответ
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}

// Структура для входящего запроса на регистрацию
type UserCreateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=driver passenger"`
}

// Структура для входящего запроса на логин
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
