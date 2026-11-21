package main

import (
	"fmt"
	"log"
	"net/http"
	"transfer-service/internal/config"
	"transfer-service/internal/db"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	// load config
	cfg := config.Load()

	// connect bd
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	// Router setup
	r := chi.NewRouter()

	// helpfull middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Простой Health Check (проверка, что сервер живой)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	// start server

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🌍 Сервер запущен на http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v\n", err)
	}

}
