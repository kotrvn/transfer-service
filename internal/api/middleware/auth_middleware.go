package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Создаем свой тип для ключа контекста, чтобы избежать конфликтов
type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UserRoleKey contextKey = "userRole"
)

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Достаем заголовок Authorization (формат: "Bearer <токен>")
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Требуется токен авторизации", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// 2. Парсим и проверяем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())

			if err != nil || !token.Valid {
				http.Error(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			// 3. Достаем данные из токена (claims)
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Ошибка чтения данных токена", http.StatusUnauthorized)
				return
			}

			userID, _ := claims["sub"].(string)
			userRole, _ := claims["role"].(string)

			id, err := uuid.Parse(userID)
			if err != nil || id == uuid.Nil || (userRole != "driver" && userRole != "passenger") {
				http.Error(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			// 4. Кладем данные в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, userRole)

			// Передаем управление дальше
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
