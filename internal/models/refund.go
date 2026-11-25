package models

type RefundJob struct {
	Intent   string
	Amount   int64
	Status   string
	Attempts int
}
