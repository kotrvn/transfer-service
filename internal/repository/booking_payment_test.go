package repository

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"transfer-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBookingPaymentPostgres(t *testing.T) {
	dsn := os.Getenv("BOOKING_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set BOOKING_TEST_DATABASE_URL to run against local PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	schema := "booking_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	exec := func(pool *pgxpool.Pool, sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(admin, "CREATE SCHEMA "+schema)
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := admin.Exec(cleanup, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	migrations, err := filepath.Glob("../../migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("migrations: %v, %v", migrations, err)
	}
	for _, path := range migrations {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		up, _, ok := strings.Cut(string(data), "-- +goose Down")
		if !ok {
			t.Fatalf("missing Down marker in %s", path)
		}
		exec(pool, up)
	}

	driver, passenger, trip := uuid.New(), uuid.New(), uuid.New()
	exec(pool, `INSERT INTO users (id, name, email, hashed_password, role)
		VALUES ($1, 'Driver', 'driver@example.com', 'hash', 'driver'),
		($2, 'Passenger', 'passenger@example.com', 'hash', 'passenger')`, driver, passenger)
	exec(pool, `INSERT INTO trips (id, driver_id, from_city, to_city, departure_at, seats_total, seats_available, price_per_seat)
		VALUES ($1, $2, 'A', 'B', NOW(), 10, 10, 12.34)`, trip, driver)
	repo := NewBookingRepository(pool)
	create := func() uuid.UUID {
		t.Helper()
		id := uuid.New()
		booking := models.Booking{ID: id, TripID: trip, PassengerID: passenger, Seats: 2, Status: "pending"}
		if err := repo.Create(ctx, &booking); err != nil || booking.CreatedAt.IsZero() {
			t.Fatalf("create: %v, %+v", err, booking)
		}
		return id
	}
	confirm := func(id uuid.UUID, intent string, amount int64, want bool) {
		t.Helper()
		if confirmed, err := repo.ConfirmPayment(ctx, id, intent, amount); err != nil || confirmed != want {
			t.Fatalf("confirm: got %t, %v, want %t", confirmed, err, want)
		}
	}

	id := create()
	if amount, err := repo.CheckoutAmount(ctx, id, passenger); err != nil || amount != 2468 {
		t.Fatalf("checkout amount: %d, %v", amount, err)
	}
	if _, err := repo.CheckoutAmount(ctx, id, driver); err != ErrBookingNotFound {
		t.Fatalf("checkout with another user: %v", err)
	}
	confirm(id, "pi_paid", 2467, false)
	confirm(id, "pi_paid", 2468, true)
	confirm(id, "pi_paid", 2468, true)
	confirm(id, "pi_duplicate", 2468, false)
	if err := repo.Cancel(ctx, id, passenger); err == nil {
		t.Fatal("paid booking was cancelled")
	}

	expired := create()
	exec(pool, "UPDATE bookings SET created_at = NOW() - interval '16 minutes' WHERE id = $1", expired)
	if _, err := repo.CheckoutAmount(ctx, expired, passenger); err != ErrBookingNotPayable {
		t.Fatalf("expired checkout: %v", err)
	}
	confirm(expired, "pi_late", 2468, false)
	if count, err := repo.CleanupExpired(ctx, models.BookingHold); err != nil || count != 1 {
		t.Fatalf("cleanup: %d, %v", count, err)
	}

	racing := create()
	type result struct {
		confirmed bool
		err       error
	}
	payment := make(chan result, 1)
	cancellation := make(chan error, 1)
	start := make(chan struct{})
	go func() {
		<-start
		confirmed, err := repo.ConfirmPayment(ctx, racing, "pi_racing", 2468)
		payment <- result{confirmed, err}
	}()
	go func() {
		<-start
		cancellation <- repo.Cancel(ctx, racing, passenger)
	}()
	close(start)
	paid, cancelErr := <-payment, <-cancellation
	if paid.err != nil || paid.confirmed == (cancelErr == nil) {
		t.Fatalf("confirmation and cancellation must have one winner: %+v, %v", paid, cancelErr)
	}
	confirm(racing, "pi_racing", 2468, paid.confirmed)
	wantSeats := 8
	if paid.confirmed {
		wantSeats -= 2
	}
	var available int
	if err := pool.QueryRow(ctx, "SELECT seats_available FROM trips WHERE id = $1", trip).Scan(&available); err != nil || available != wantSeats {
		t.Fatalf("available seats: %d, %v, want %d", available, err, wantSeats)
	}
	bookings, err := repo.GetByPassengerID(ctx, passenger)
	if err != nil || len(bookings) != 3 {
		t.Fatalf("list: %v, %v", bookings, err)
	}
	for _, booking := range bookings {
		if booking.PaymentAvailable || booking.Paid != (booking.Status == "confirmed") {
			t.Errorf("unexpected payment flags: %+v", booking)
		}
	}

	pending := create()
	tripRepo := NewTripRepository(pool)
	if err := tripRepo.Complete(ctx, trip, passenger); err != ErrTripNotFound {
		t.Fatalf("complete another driver's trip: %v", err)
	}
	exec(pool, "UPDATE trips SET departure_at = NOW() + interval '1 hour' WHERE id = $1", trip)
	if err := tripRepo.Complete(ctx, trip, driver); err != ErrTripNotDeparted {
		t.Fatalf("complete before departure: %v", err)
	}
	exec(pool, "UPDATE trips SET departure_at = NOW() - interval '1 hour' WHERE id = $1", trip)
	for i := 0; i < 2; i++ {
		if err := tripRepo.Complete(ctx, trip, driver); err != nil {
			t.Fatalf("complete trip: %v", err)
		}
	}
	if err := repo.Create(ctx, &models.Booking{ID: uuid.New(), TripID: trip, PassengerID: passenger, Seats: 1, Status: "pending"}); err != ErrTripCompleted {
		t.Fatalf("book completed trip: %v", err)
	}
	if _, err := repo.CheckoutAmount(ctx, pending, passenger); err != ErrBookingNotPayable {
		t.Fatalf("pay for completed trip: %v", err)
	}
	if err := repo.Cancel(ctx, pending, passenger); err == nil {
		t.Fatal("cancelled pending booking after trip completion")
	}
	confirm(pending, "pi_after_trip", 2468, false)
	confirm(id, "pi_paid", 2468, true)
	bookings, err = repo.GetByPassengerID(ctx, passenger)
	if err != nil {
		t.Fatal(err)
	}
	statuses := make(map[uuid.UUID]string)
	for _, booking := range bookings {
		statuses[booking.ID] = booking.Status
	}
	if statuses[id] != "completed" || statuses[pending] != "expired" || statuses[expired] != "expired" {
		t.Fatalf("statuses after completion: %v", statuses)
	}
	if count, err := repo.CleanupExpired(ctx, models.BookingHold); err != nil || count != 1 {
		t.Fatalf("cleanup after trip completion: %d, %v", count, err)
	}
	if count, err := repo.CleanupExpired(ctx, models.BookingHold); err != nil || count != 0 {
		t.Fatalf("repeated cleanup after trip completion: %d, %v", count, err)
	}
	trips, err := tripRepo.List(ctx, "", "")
	if err != nil || len(trips) != 1 || trips[0].CompletedAt == nil || trips[0].SeatsAvailable != wantSeats {
		t.Fatalf("completed trip result: %+v, %v", trips, err)
	}
}
