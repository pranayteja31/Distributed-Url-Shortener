package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"pranayteja31/Urlshortener/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestURLRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	url := &models.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CreatedAt:   createdAt,
		ExpiresAt:   &expiresAt,
		ClickCount:  0,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO urls(
			short_code,
			original_url,
			created_at,
			expires_at,
			click_count
		)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id;
	`)).
		WithArgs(
			url.ShortCode,
			url.OriginalURL,
			url.CreatedAt,
			url.ExpiresAt,
			url.ClickCount,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	err = repo.Create(url)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url.ID != 1 {
		t.Errorf("expected ID 1, got %d", url.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestURLRepository_FindByShortCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		1,
		"abc123",
		"https://example.com",
		createdAt,
		expiresAt,
		5,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=$1;`,
	)).
		WithArgs("abc123").
		WillReturnRows(rows)

	url, err := repo.FindByShortCode("abc123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url == nil {
		t.Fatal("expected URL, got nil")
	}

	if url.ID != 1 {
		t.Errorf("expected ID 1, got %d", url.ID)
	}

	if url.ShortCode != "abc123" {
		t.Errorf("expected short code abc123, got %s", url.ShortCode)
	}

	if url.OriginalURL != "https://example.com" {
		t.Errorf("unexpected original URL: %s", url.OriginalURL)
	}

	if url.ClickCount != 5 {
		t.Errorf("expected click count 5, got %d", url.ClickCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		10,
		"xyz789",
		"https://example.com",
		createdAt,
		expiresAt,
		3,
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=$1;`,
	)).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	url, err := repo.FindByID(10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url == nil {
		t.Fatal("expected URL, got nil")
	}

	if url.ID != 10 {
		t.Errorf("expected ID 10, got %d", url.ID)
	}

	if url.ShortCode != "xyz789" {
		t.Errorf("expected short code xyz789, got %s", url.ShortCode)
	}

	if url.ClickCount != 3 {
		t.Errorf("expected click count 3, got %d", url.ClickCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=$1;`,
	)).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	url, err := repo.FindByID(999)

	if url != nil {
		t.Fatal("expected nil URL")
	}

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	expiresAt := time.Now().Add(24 * time.Hour)

	url := &models.URL{
		ID:          10,
		OriginalURL: "https://updated-example.com",
		ExpiresAt:   &expiresAt,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE urls
		SET
			original_url = $1,
			expires_at = $2
		WHERE id = $3;
	`)).
		WithArgs(
			url.OriginalURL,
			url.ExpiresAt,
			url.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(url)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM urls
		WHERE id = $1;
	`)).
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLRepository_IncrementCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE urls SET click_count = click_count + 1 WHERE id=$1;`,
	)).
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.IncrementCount(10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).
		AddRow(
			1,
			"abc123",
			"https://example.com",
			createdAt,
			expiresAt,
			5,
		).
		AddRow(
			2,
			"xyz789",
			"https://google.com",
			createdAt,
			expiresAt,
			10,
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, short_code, original_url, created_at, expires_at, click_count
		FROM urls
		ORDER BY created_at DESC;
	`)).
		WillReturnRows(rows)

	urls, err := repo.List()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}

	if urls[0].ShortCode != "abc123" {
		t.Errorf("expected abc123, got %s", urls[0].ShortCode)
	}

	if urls[1].ShortCode != "xyz789" {
		t.Errorf("expected xyz789, got %s", urls[1].ShortCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}