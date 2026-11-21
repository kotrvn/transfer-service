package repository

import (
	"context"
	"errors"
	"transfer-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Создание пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, name, email, hashed_password, role) 
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Name, user.Email, user.HashedPassword, user.Role)
	return err
}

// Поиск пользователя по email (для логина)
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, hashed_password, role FROM users WHERE email = $1`

	var user models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.Role,
	)

	if err != nil {
		return nil, errors.New("Пользователь не найден")
	}
	return &user, nil
}
