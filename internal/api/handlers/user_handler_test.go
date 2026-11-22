package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"transfer-service/internal/models"

	"github.com/stretchr/testify/assert"
)

// 1. Создаем Фейковый (Mock) репозиторий
type mockUserRepo struct {
	// Здесь мы можем задать поведение для базы данных
	fakeUsers map[string]*models.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	if m.fakeUsers == nil {
		m.fakeUsers = make(map[string]*models.User)
	}
	m.fakeUsers[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, ok := m.fakeUsers[email]; ok {
		return user, nil
	}
	return nil, assert.AnError // Возвращаем ошибку, если юзера нет
}

// 2. Сам тест
func TestRegisterHandler_Success(t *testing.T) {
	// Подключаем фейковый репозиторий
	mockRepo := &mockUserRepo{}
	handler := NewUserHandler(mockRepo, "test_secret")

	// Готовим JSON для отправки
	reqBody := `{"name": "Тест", "email": "test@example.com", "password": "123456", "role": "driver"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users/", bytes.NewBufferString(reqBody))

	// Создаем "фейковый" ResponseWriter, чтобы перехватить ответ
	rr := httptest.NewRecorder()

	// Вызываем наш хендлер напрямую!
	handler.Register(rr, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusCreated, rr.Code) // Должен быть код 201

	// Проверяем, что юзер реально попал в "фейковую" базу
	var createdUser models.User
	json.NewDecoder(rr.Body).Decode(&createdUser)
	assert.Equal(t, "test@example.com", createdUser.Email)
	assert.Equal(t, "driver", createdUser.Role)
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	mockRepo := &mockUserRepo{}
	handler := NewUserHandler(mockRepo, "test_secret")

	// Отправляем НЕВАЛИДНЫЙ JSON (короткий пароль)
	reqBody := `{"name": "Тест", "email": "test@example.com", "password": "123", "role": "driver"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users/", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	// Проверяем, что сервер вернул 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "должно быть не короче 6") // Проверяем текст ошибки
}
