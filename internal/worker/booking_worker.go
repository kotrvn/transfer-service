package worker

import (
	"context"
	"log"
	"time"
)

// Интерфейс — говорит воркеру, что ему нужно от "уборщика"
type BookingCleaner interface {
	CleanupExpired(ctx context.Context, expireAfter time.Duration) (int, error)
}

func StartBookingCleanupWorker(ctx context.Context, cleaner BookingCleaner, interval time.Duration, expireAfter time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("🧹 Воркер очистки запущен (проверка каждые %v)", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Воркер остановлен")
			return
		case <-ticker.C:
			count, err := cleaner.CleanupExpired(ctx, expireAfter)
			if err != nil {
				log.Printf("Воркер: ошибка: %v", err)
				continue
			}
			if count > 0 {
				log.Printf("🧹 Воркер: закрыто броней с возвратом мест: %d", count)
			}
		}
	}
}
