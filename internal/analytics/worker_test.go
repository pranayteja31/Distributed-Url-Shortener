package analytics

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
)

// ============================================================
// HELPER
// ============================================================

func setupAnalyticsWorker(t *testing.T, bufferSize int) (
	*Worker,
	sqlmock.Sqlmock,
	func(),
) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	clickRepo := repository.NewClickRepository(sqlxDB)

	worker := NewWorker(
		clickRepo,
		bufferSize,
	)

	cleanup := func() {
		db.Close()
	}

	return worker, mock, cleanup
}

// ============================================================
// RECORD - QUEUES EVENT
// ============================================================

func TestWorker_Record(t *testing.T) {
	worker, _, cleanup := setupAnalyticsWorker(t, 1)
	defer cleanup()

	click := models.URLClick{
		URLID:     10,
		ClickedAt: time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
		Referrer:  "https://google.com",
	}

	worker.Record(click)

	select {
	case queuedClick := <-worker.events:

		if queuedClick.URLID != click.URLID {
			t.Fatalf(
				"expected URL ID %d, got %d",
				click.URLID,
				queuedClick.URLID,
			)
		}

		if queuedClick.IPAddress != click.IPAddress {
			t.Fatalf(
				"expected IP %s, got %s",
				click.IPAddress,
				queuedClick.IPAddress,
			)
		}

		if queuedClick.UserAgent != click.UserAgent {
			t.Fatalf(
				"expected User-Agent %s, got %s",
				click.UserAgent,
				queuedClick.UserAgent,
			)
		}

		if queuedClick.Referrer != click.Referrer {
			t.Fatalf(
				"expected Referrer %s, got %s",
				click.Referrer,
				queuedClick.Referrer,
			)
		}

	default:
		t.Fatal("expected click event to be queued")
	}
}

// ============================================================
// RECORD - QUEUE FULL
// ============================================================

func TestWorker_Record_QueueFull(t *testing.T) {
	worker, _, cleanup := setupAnalyticsWorker(t, 1)
	defer cleanup()

	click1 := models.URLClick{
		URLID: 1,
	}

	click2 := models.URLClick{
		URLID: 2,
	}

	// Fill the queue.
	worker.Record(click1)

	// This should return immediately instead of blocking.
	done := make(chan struct{})

	go func() {
		worker.Record(click2)
		close(done)
	}()

	select {
	case <-done:
		// Expected.
	case <-time.After(time.Second):
		t.Fatal("Record blocked when queue was full")
	}

	// Only the first event should remain.
	if len(worker.events) != 1 {
		t.Fatalf(
			"expected queue length 1, got %d",
			len(worker.events),
		)
	}

	queuedClick := <-worker.events

	if queuedClick.URLID != 1 {
		t.Fatalf(
			"expected first event to remain queued, got URL ID %d",
			queuedClick.URLID,
		)
	}
}

// ============================================================
// START - PROCESSES EVENT
// ============================================================

func TestWorker_Start_ProcessesEvent(t *testing.T) {
	worker, mock, cleanup := setupAnalyticsWorker(t, 1)
	defer cleanup()

	click := models.URLClick{
		URLID:     10,
		ClickedAt: time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
		Referrer:  "https://google.com",
	}

	// ClickRepository uses QueryRowx because the SQL contains
	// RETURNING id, so sqlmock must use ExpectQuery.
	mock.ExpectQuery(
		`INSERT INTO url_clicks`,
	).
		WithArgs(
			click.URLID,
			click.ClickedAt,
			click.IPAddress,
			click.UserAgent,
			click.Referrer,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(1),
		)

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Start(ctx)

	worker.Record(click)

	// Wait until the database expectation is fulfilled.
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if err := mock.ExpectationsWereMet(); err == nil {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	worker.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"worker did not process event: %v",
			err,
		)
	}
}

// ============================================================
// START - DATABASE ERROR
// ============================================================

func TestWorker_Start_DatabaseError(t *testing.T) {
	worker, mock, cleanup := setupAnalyticsWorker(t, 1)
	defer cleanup()

	click := models.URLClick{
		URLID:     20,
		ClickedAt: time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
		Referrer:  "",
	}

	mock.ExpectQuery(
		`INSERT INTO url_clicks`,
	).
		WithArgs(
			click.URLID,
			click.ClickedAt,
			click.IPAddress,
			click.UserAgent,
			click.Referrer,
		).
		WillReturnError(sql.ErrConnDone)

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Start(ctx)

	worker.Record(click)

	// Wait until the query is executed.
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if err := mock.ExpectationsWereMet(); err == nil {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	worker.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expected database operation to occur: %v",
			err,
		)
	}
}

// ============================================================
// SHUTDOWN - DRAINS QUEUED EVENTS
// ============================================================

func TestWorker_Shutdown_DrainsQueue(t *testing.T) {
	worker, mock, cleanup := setupAnalyticsWorker(t, 2)
	defer cleanup()

	click1 := models.URLClick{
		URLID:     100,
		ClickedAt: time.Now(),
		IPAddress: "10.0.0.1",
		UserAgent: "agent-1",
		Referrer:  "referrer-1",
	}

	click2 := models.URLClick{
		URLID:     200,
		ClickedAt: time.Now(),
		IPAddress: "10.0.0.2",
		UserAgent: "agent-2",
		Referrer:  "referrer-2",
	}

	// The worker processes click1.
	mock.ExpectQuery(
		`INSERT INTO url_clicks`,
	).
		WithArgs(
			click1.URLID,
			click1.ClickedAt,
			click1.IPAddress,
			click1.UserAgent,
			click1.Referrer,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(1),
		)

	// The worker processes click2.
	mock.ExpectQuery(
		`INSERT INTO url_clicks`,
	).
		WithArgs(
			click2.URLID,
			click2.ClickedAt,
			click2.IPAddress,
			click2.UserAgent,
			click2.Referrer,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(2),
		)

	// Queue both events before shutdown.
	worker.Record(click1)
	worker.Record(click2)

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Start(ctx)

	// Allow worker to start.
	time.Sleep(20 * time.Millisecond)

	// Request shutdown.
	cancel()

	// Start() should drain remaining queued events
	// before returning.
	worker.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"expected queued events to be drained: %v",
			err,
		)
	}
}

// ============================================================
// WAIT - WORKER SHUTDOWN
// ============================================================

func TestWorker_Wait(t *testing.T) {
	worker, _, cleanup := setupAnalyticsWorker(t, 1)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Start(ctx)

	// Give Start() a chance to register itself
	// with the WaitGroup.
	time.Sleep(20 * time.Millisecond)

	cancel()

	done := make(chan struct{})

	go func() {
		worker.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected.
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not shut down")
	}
}
