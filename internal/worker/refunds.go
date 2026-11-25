package worker

import (
	"context"
	"log"
	"time"

	"transfer-service/internal/models"
)

type RefundQueue interface {
	ClaimRefund(context.Context) (*models.RefundJob, error)
	UpdateRefund(context.Context, models.RefundJob, string, string, time.Duration) error
}

type RefundProvider interface {
	ReconcileRefund(context.Context, models.RefundJob) (string, error)
}

func RunRefundWorker(ctx context.Context, queue RefundQueue, provider RefundProvider) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		processRefund(ctx, queue, provider)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func processRefund(ctx context.Context, queue RefundQueue, provider RefundProvider) {
	jobCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	job, err := queue.ClaimRefund(jobCtx)
	if err != nil {
		log.Printf("Очередь возвратов: %v", err)
		return
	}
	if job == nil {
		return
	}
	status, err := provider.ReconcileRefund(jobCtx, *job)
	delay := time.Minute
	lastError := ""
	switch {
	case err != nil:
		status = job.Status
		lastError = err.Error()
		delay = time.Duration(min(job.Attempts, 12)) * time.Minute
	case status == "failed" || status == "requires_action":
		log.Printf("Возврат Stripe %s требует проверки: %s", job.Intent, status)
	}
	if err := queue.UpdateRefund(jobCtx, *job, status, lastError, delay); err != nil {
		log.Printf("Сохранение результата возврата Stripe %s: %v", job.Intent, err)
	}
}
