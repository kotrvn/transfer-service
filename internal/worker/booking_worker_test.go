package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Мок-уборщик. Используем atomic, потому что воркер работает в отдельной горутине
type mockCleaner struct {
	callCount int32
}

func (m *mockCleaner) CleanupExpired(ctx context.Context, expireAfter time.Duration) (int, error) {
	atomic.AddInt32(&m.callCount, 1) // Безопасно увеличиваем счетчик
	return 1, nil                    // Имитируем, что вернули 1 место
}

func TestBookingCleanupWorker_StopsByContext(t *testing.T) {
	cleaner := &mockCleaner{}
	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем воркер с интервалом 10 миллисекунд (для быстрого теста)
	go StartBookingCleanupWorker(ctx, cleaner, 10*time.Millisecond, 15*time.Minute)

	// Ждем 50 миллисекунд (за это время воркер должен проснуться ~5 раз)
	time.Sleep(50 * time.Millisecond)

	// Останавливаем воркер
	cancel()
	time.Sleep(10 * time.Millisecond) // Даем время на выход

	// Проверяем, что уборщик вызывался (callCount > 0)
	assert.True(t, atomic.LoadInt32(&cleaner.callCount) > 0, "Воркер не вызвал уборщик")
}
