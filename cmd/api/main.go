package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"transfer-service/internal/api/handlers"
	"transfer-service/internal/api/middleware"
	"transfer-service/internal/config"
	"transfer-service/internal/db"
	"transfer-service/internal/repository"
	"transfer-service/internal/worker"

	_ "transfer-service/docs" // Важно: импортируем сгенерированные доки

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Transfer Service API
// @version 1.0
// @description API для поиска и бронирования трансферов.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Введите "Bearer " (с пробелом) и затем ваш JWT токен.
func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	// Инициализация репозиториев и хендлеров
	userRepo := repository.NewUserRepository(database.Pool)
	tripRepo := repository.NewTripRepository(database.Pool)
	bookingRepo := repository.NewBookingRepository(database.Pool)

	userHandler := handlers.NewUserHandler(userRepo, cfg.JWTSecret)
	tripHandler := handlers.NewTripHandler(tripRepo)
	bookingHandler := handlers.NewBookingHandler(bookingRepo)

	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // URL к swagger.json
	))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Открытые маршруты
		r.Post("/users/", userHandler.Register)
		r.Post("/users/login", userHandler.Login)

		// Публичный список поездок (пассажиры тоже должны их видеть)
		r.Get("/trips", tripHandler.List)

		// Защищенные маршруты (только с валидным токеном)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))
			r.Post("/trips", tripHandler.Create)
			r.Post("/bookings", bookingHandler.Create)
			r.Get("/bookings/my", bookingHandler.ListMy)
			r.Delete("/bookings/{id}", bookingHandler.Cancel)
		})
	})

	// --- ЗАПУСК ВОРКЕРА ---
	// Создаем контекст, который отменится при нажатии Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запускаем воркер в отдельной горутине (он не блокирует сервер)
	go worker.StartBookingCleanupWorker(ctx, bookingRepo, 1*time.Minute, 15*time.Minute)

	// --- ЗАПУСК СЕРВЕРА ---
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🌍 Сервер запущен на http://localhost%s\n", addr)

	// Запускаем сервер в горутине, чтобы он не блокировал ожидание Ctrl+C
	go func() {
		if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Ждем сигнала на выключение (Ctrl+C)
	<-ctx.Done()
	log.Println("Получен сигнал завершения, выключаем сервер...")

}
