package handlers

import (
	"encoding/json"
	"net/http"
	"transfer-service/internal/auth"
	"transfer-service/internal/models"
	"transfer-service/internal/repository"
	"uuid"
)

type UserHandler struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserHandler(repo *repository.UserRepository, secret string) *UserHandler {
	return &UserHandler{repo: repo, jwtSecret: secret}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя (водителя или пассажира).
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.UserCreateRequest true "Данные пользователя"
// @Success 201 {object} models.User
// @Failure 400,409 {string} string
// @Router /users/ [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest
	if err := decodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//pass hash
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// create user
	user := &models.User{
		ID:             uuid.New(),
		Name:           req.Name,
		Email:          req.Email,
		HashedPassword: hashedPassword,
		Role:           req.Role,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		http.Error(w, "Email уже занят или ошибка БД", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login godoc
// @Summary Вход в систему
// @Description Возвращает JWT токен для доступа к защищенным маршрутам.
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.UserLoginRequest true "Email и пароль"
// @Success 200 {object} map[string]string "Пример: {"token": "eyJ..."}"
// @Failure 401 {string} string
// @Router /users/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := decodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// token generate
	token, err := auth.GenerateJWT(user.ID.String(), user.Role, h.jwtSecret)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
