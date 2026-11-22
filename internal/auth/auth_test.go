package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	password := "super_secret_password"

	// 1. Хешируем
	hash, err := HashPassword(password)
	assert.NoError(t, err) // Проверяем, что ошибки нет
	assert.NotEmpty(t, hash)

	// 2. Проверяем правильный пароль
	assert.True(t, CheckPasswordHash(password, hash))

	// 3. Проверяем НЕправильный пароль
	assert.False(t, CheckPasswordHash("wrong_password", hash))
}

func TestGenerateJWT(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	role := "driver"
	secret := "my_test_secret"

	token, err := GenerateJWT(userID, role, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
