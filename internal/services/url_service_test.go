package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)
type mockCache struct {
	deleteCalled bool
	deletedKey   string

	getData  []byte
	getFound bool
	getErr   error

	setCalled bool
}

func (m *mockCache) Get(
	ctx context.Context,
	key string,
) ([]byte, bool, error) {
	return m.getData, m.getFound, m.getErr
}

func (m *mockCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	ttl time.Duration,
) error {
	m.setCalled = true
	return nil
}

func (m *mockCache) Delete(
	ctx context.Context,
	key string,
) error {
	m.deleteCalled = true
	m.deletedKey = key
	return nil
}
func TestURLServices_GetURL_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	// GetURL does not use Redis or Analytics Worker,
	// so nil is safe for this test.
	service := NewURLServices(repo, nil, nil)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=\$1;`,
	).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	url, err := service.GetURL(999)

	if url != nil {
		t.Fatal("expected nil URL")
	}

	if !errors.Is(err, ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_GetURL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	service := NewURLServices(repo, nil, nil)

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

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=\$1;`,
	).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	url, err := service.GetURL(1)

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
		t.Errorf("expected https://example.com, got %s", url.OriginalURL)
	}

	if url.ClickCount != 5 {
		t.Errorf("expected click count 5, got %d", url.ClickCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_DeleteURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	// Find the URL before deleting it.
	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		10,
		"abc123",
		"https://example.com",
		createdAt,
		expiresAt,
		5,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=\$1;`,
	).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	// Delete the URL.
	mock.ExpectExec(
		`DELETE FROM urls\s+WHERE id = \$1;`,
	).
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = service.DeleteURL(10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify Redis invalidation happened.
	if !mockCache.deleteCalled {
		t.Fatal("expected cache Delete to be called")
	}

	if mockCache.deletedKey != "url:abc123" {
		t.Errorf(
			"expected deleted key url:abc123, got %s",
			mockCache.deletedKey,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_DeleteURL_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=\$1;`,
	).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err = service.DeleteURL(999)

	if !errors.Is(err, ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound, got %v", err)
	}

	// Cache should NOT be touched when the URL doesn't exist.
	if mockCache.deleteCalled {
		t.Fatal("cache Delete should not be called for a missing URL")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_UpdateURL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	createdAt := time.Now()
	oldExpiry := createdAt.Add(24 * time.Hour)

	// Existing URL returned by FindByID.
	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		10,
		"abc123",
		"https://old-example.com",
		createdAt,
		oldExpiry,
		5,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=\$1;`,
	).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	// Updated URL.
	mock.ExpectExec(
		`UPDATE urls\s+SET\s+original_url = \$1,\s+expires_at = \$2\s+WHERE id = \$3;`,
	).
		WithArgs(
			"https://example.com",
			sqlmock.AnyArg(),
			int64(10),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	url := &models.URL{
		ID:          10,
		OriginalURL: "https://example.com",
	}

	updatedURL, err := service.UpdateURL(url, 30)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedURL == nil {
		t.Fatal("expected updated URL, got nil")
	}

	if updatedURL.ID != 10 {
		t.Errorf("expected ID 10, got %d", updatedURL.ID)
	}

	if updatedURL.ShortCode != "abc123" {
		t.Errorf("expected short code abc123, got %s", updatedURL.ShortCode)
	}

	if updatedURL.OriginalURL != "https://example.com" {
		t.Errorf(
			"expected https://example.com, got %s",
			updatedURL.OriginalURL,
		)
	}

	// Cache must be invalidated after updating the URL.
	if !mockCache.deleteCalled {
		t.Fatal("expected cache Delete to be called")
	}

	if mockCache.deletedKey != "url:abc123" {
		t.Errorf(
			"expected deleted key url:abc123, got %s",
			mockCache.deletedKey,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_UpdateURL_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(repo, mockCache, nil)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE id=\$1;`,
	).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	url := &models.URL{
		ID:          999,
		OriginalURL: "https://example.com",
	}

	updatedURL, err := service.UpdateURL(url, 30)

	if updatedURL != nil {
		t.Fatal("expected nil URL")
	}

	if !errors.Is(err, ErrURLNotFound) {
		t.Fatalf("expected ErrURLNotFound, got %v", err)
	}

	if mockCache.deleteCalled {
		t.Fatal("cache Delete should not be called")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_UpdateURL_InvalidURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(repo, mockCache, nil)

	url := &models.URL{
		ID:          10,
		OriginalURL: "not-a-valid-url",
	}

	updatedURL, err := service.UpdateURL(url, 30)

	if updatedURL != nil {
		t.Fatal("expected nil URL")
	}

	if err == nil {
		t.Fatal("expected validation error")
	}

	// Database should never be queried because validation
	// happens before FindByID.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database operation: %v", err)
	}

	if mockCache.deleteCalled {
		t.Fatal("cache Delete should not be called")
	}
}
func TestURLServices_CreateShortURL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(repo, mockCache, nil)

	// First generated short-code does not exist.
	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	// Create URL.
	mock.ExpectQuery(
	`(?s)INSERT INTO urls\(\s*short_code,\s*original_url,\s*created_at,\s*expires_at,\s*click_count\s*\)\s*VALUES \(\$1,\$2,\$3,\$4,\$5\)\s*RETURNING id;`,
).
	WithArgs(
		sqlmock.AnyArg(),
		"https://example.com",
		sqlmock.AnyArg(),
		sqlmock.AnyArg(),
		0,
	).
	WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(20),
	)
	url, err := service.CreateShortURL(
		"https://example.com",
		30,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url == nil {
		t.Fatal("expected URL, got nil")
	}

	if url.ID != 20 {
		t.Errorf("expected ID 20, got %d", url.ID)
	}

	if url.OriginalURL != "https://example.com" {
		t.Errorf(
			"expected https://example.com, got %s",
			url.OriginalURL,
		)
	}

	if url.ShortCode == "" {
		t.Fatal("expected generated short code")
	}

	if len(url.ShortCode) != 6 {
		t.Errorf(
			"expected short code length 6, got %d",
			len(url.ShortCode),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_CreateShortURL_InvalidURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(repo, mockCache, nil)

	url, err := service.CreateShortURL(
		"not-a-valid-url",
		30,
	)

	if url != nil {
		t.Fatal("expected nil URL")
	}

	if err == nil {
		t.Fatal("expected validation error")
	}

	// Validation happens before any database operation.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database operation: %v", err)
	}
}
func TestURLServices_CreateShortURL_DuplicateShortCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{}

	service := NewURLServices(repo, mockCache, nil)

	// First generated short code already exists.
	existingRows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		1,
		"existing",
		"https://old-example.com",
		time.Now(),
		time.Now().Add(24*time.Hour),
		5,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(existingRows)

	// Second generated short code is available.
	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	// Create using the second code.
	mock.ExpectQuery(
		`(?s)INSERT INTO urls\(\s*short_code,\s*original_url,\s*created_at,\s*expires_at,\s*click_count\s*\)\s*VALUES \(\$1,\$2,\$3,\$4,\$5\)\s*RETURNING id;`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			"https://example.com",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			0,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(20),
		)

	url, err := service.CreateShortURL(
		"https://example.com",
		30,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url == nil {
		t.Fatal("expected URL, got nil")
	}

	if url.ID != 20 {
		t.Errorf("expected ID 20, got %d", url.ID)
	}

	if url.ShortCode == "" {
		t.Fatal("expected generated short code")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_RedirectURL_CacheHit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	cachedURL := models.URL{
		ID:          10,
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CreatedAt:   createdAt,
		ExpiresAt:   &expiresAt,
		ClickCount:  5,
	}

	cacheData, err := json.Marshal(cachedURL)
	if err != nil {
		t.Fatalf("failed to marshal URL: %v", err)
	}

	mockCache := &mockCache{
		getData:  cacheData,
		getFound: true,
		getErr:   nil,
	}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	// Cache hit should NOT query PostgreSQL for URL details.
	mock.ExpectExec(
		`UPDATE urls SET click_count = click_count \+ 1 WHERE id=\$1;`,
	).
		WithArgs(int64(10)).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	originalURL, err := service.RedirectURL(
		"abc123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if originalURL != "https://example.com" {
		t.Errorf(
			"expected https://example.com, got %s",
			originalURL,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_RedirectURL_CacheMiss(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	// Redis cache MISS.
	mockCache := &mockCache{
		getData:  nil,
		getFound: false,
		getErr:   nil,
	}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	// PostgreSQL lookup.
	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		10,
		"abc123",
		"https://example.com",
		createdAt,
		expiresAt,
		5,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs("abc123").
		WillReturnRows(rows)

	// Increment click count.
	mock.ExpectExec(
		`UPDATE urls SET click_count = click_count \+ 1 WHERE id=\$1;`,
	).
		WithArgs(int64(10)).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	originalURL, err := service.RedirectURL(
		"abc123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if originalURL != "https://example.com" {
		t.Errorf(
			"expected https://example.com, got %s",
			originalURL,
		)
	}

	// URL should be cached after PostgreSQL lookup.
	if !mockCache.setCalled {
		t.Fatal("expected cache Set to be called")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_RedirectURL_ExpiredCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	createdAt := time.Now().Add(-48 * time.Hour)
	expiresAt := time.Now().Add(-1 * time.Hour)

	cachedURL := models.URL{
		ID:          10,
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CreatedAt:   createdAt,
		ExpiresAt:   &expiresAt,
		ClickCount:  5,
	}

	cacheData, err := json.Marshal(cachedURL)
	if err != nil {
		t.Fatalf("failed to marshal URL: %v", err)
	}

	mockCache := &mockCache{
		getData:  cacheData,
		getFound: true,
		getErr:   nil,
	}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	originalURL, err := service.RedirectURL(
		"abc123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if originalURL != "" {
		t.Errorf(
			"expected empty URL, got %s",
			originalURL,
		)
	}

	if !errors.Is(err, ErrURLExpired) {
		t.Fatalf(
			"expected ErrURLExpired, got %v",
			err,
		)
	}

	// Expired cache entry should be deleted.
	if !mockCache.deleteCalled {
		t.Fatal("expected expired cache entry to be deleted")
	}

	if mockCache.deletedKey != "url:abc123" {
		t.Errorf(
			"expected deleted key url:abc123, got %s",
			mockCache.deletedKey,
		)
	}

	// No PostgreSQL query should happen because
	// the expired URL was already found in Redis.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database operation: %v", err)
	}
}
func TestURLServices_RedirectURL_ExpiredPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	// Redis MISS.
	mockCache := &mockCache{
		getData:  nil,
		getFound: false,
		getErr:   nil,
	}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	createdAt := time.Now().Add(-48 * time.Hour)
	expiresAt := time.Now().Add(-1 * time.Hour)

	// PostgreSQL contains an expired URL.
	rows := sqlmock.NewRows([]string{
		"id",
		"short_code",
		"original_url",
		"created_at",
		"expires_at",
		"click_count",
	}).AddRow(
		10,
		"abc123",
		"https://example.com",
		createdAt,
		expiresAt,
		5,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs("abc123").
		WillReturnRows(rows)

	originalURL, err := service.RedirectURL(
		"abc123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if originalURL != "" {
		t.Errorf(
			"expected empty URL, got %s",
			originalURL,
		)
	}

	if !errors.Is(err, ErrURLExpired) {
		t.Fatalf(
			"expected ErrURLExpired, got %v",
			err,
		)
	}

	// Expired URL must not increment clicks.
	// It must also not be cached.
	if mockCache.setCalled {
		t.Fatal("expired URL should not be cached")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestURLServices_RedirectURL_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{
		getData:  nil,
		getFound: false,
		getErr:   nil,
	}

	service := NewURLServices(
		repo,
		mockCache,
		nil,
	)

	// Redis MISS → PostgreSQL lookup.
	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs("missing123").
		WillReturnError(sql.ErrNoRows)

	originalURL, err := service.RedirectURL(
		"missing123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if originalURL != "" {
		t.Errorf(
			"expected empty URL, got %s",
			originalURL,
		)
	}

	if !errors.Is(err, ErrURLNotFound) {
		t.Fatalf(
			"expected ErrURLNotFound, got %v",
			err,
		)
	}

	// No cache SET for a URL that doesn't exist.
	if mockCache.setCalled {
		t.Fatal("cache Set should not be called")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
// ============================================================
// REDIRECT URL - CACHE ERROR
// ============================================================

func TestURLServices_RedirectURL_CacheError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	mockCache := &mockCache{
		getData:  nil,
		getFound: false,
		getErr:   errors.New("redis unavailable"),
	}

	service := NewURLServices(repo, mockCache, nil)

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
		"abc123",
		"https://example.com",
		createdAt,
		expiresAt,
		5,
	)

	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs("abc123").
		WillReturnRows(rows)

	mock.ExpectExec(
		`UPDATE urls SET click_count = click_count \+ 1 WHERE id=\$1;`,
	).
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	originalURL, err := service.RedirectURL(
		"abc123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if originalURL != "https://example.com" {
		t.Errorf(
			"expected https://example.com, got %s",
			originalURL,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================
// REDIRECT URL - CLICK COUNT ERROR
// ============================================================

func TestURLServices_RedirectURL_IncrementCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour)

	cachedURL := models.URL{
		ID:          10,
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CreatedAt:   createdAt,
		ExpiresAt:   &expiresAt,
		ClickCount:  5,
	}

	cacheData, err := json.Marshal(cachedURL)
	if err != nil {
		t.Fatalf("failed to marshal URL: %v", err)
	}

	mockCache := &mockCache{
		getData:  cacheData,
		getFound: true,
	}

	service := NewURLServices(repo, mockCache, nil)

	mock.ExpectExec(
		`UPDATE urls SET click_count = click_count \+ 1 WHERE id=\$1;`,
	).
		WithArgs(int64(10)).
		WillReturnError(errors.New("database unavailable"))

	originalURL, err := service.RedirectURL(
		"abc123",
		"127.0.0.1",
		"test-agent",
		"",
	)

	if originalURL != "" {
		t.Errorf(
			"expected empty URL, got %s",
			originalURL,
		)
	}

	if err == nil {
		t.Fatal("expected increment count error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================
// GET URL - SUCCESS
// ============================================================



// ============================================================
// GET URL - NOT FOUND
// ============================================================


// ============================================================
// LIST URLS - SUCCESS
// ============================================================

func TestURLServices_ListURLs_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	service := NewURLServices(repo, &mockCache{}, nil)

	now := time.Now()

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
			now,
			now.Add(24*time.Hour),
			5,
		).
		AddRow(
			2,
			"xyz789",
			"https://google.com",
			now,
			now.Add(48*time.Hour),
			10,
		)

	mock.ExpectQuery(
		`SELECT id, short_code, original_url, created_at, expires_at, click_count\s+FROM urls\s+ORDER BY created_at DESC;`,
	).
		WillReturnRows(rows)

	urls, err := service.ListURLs()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}

	if urls[0].ShortCode != "abc123" {
		t.Errorf(
			"expected abc123, got %s",
			urls[0].ShortCode,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================
// LIST URLS - DATABASE ERROR
// ============================================================

func TestURLServices_ListURLs_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	service := NewURLServices(repo, &mockCache{}, nil)

	mock.ExpectQuery(
		`SELECT id, short_code, original_url, created_at, expires_at, click_count\s+FROM urls\s+ORDER BY created_at DESC;`,
	).
		WillReturnError(errors.New("database unavailable"))

	urls, err := service.ListURLs()

	if urls != nil {
		t.Fatal("expected nil URLs")
	}

	if err == nil {
		t.Fatal("expected database error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================
// CREATE SHORT URL - DATABASE ERROR
// ============================================================

func TestURLServices_CreateShortURL_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := repository.NewRepository(sqlxDB)

	service := NewURLServices(repo, &mockCache{}, nil)

	// Short code available.
	mock.ExpectQuery(
		`SELECT id,short_code,original_url,created_at,expires_at,click_count FROM urls WHERE short_code=\$1;`,
	).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	// Database insert fails.
	mock.ExpectQuery(
		`(?s)INSERT INTO urls\(\s*short_code,\s*original_url,\s*created_at,\s*expires_at,\s*click_count\s*\)\s*VALUES \(\$1,\$2,\$3,\$4,\$5\)\s*RETURNING id;`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			"https://example.com",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			0,
		).
		WillReturnError(errors.New("database unavailable"))

	url, err := service.CreateShortURL(
		"https://example.com",
		30,
	)

	if url != nil {
		t.Fatal("expected nil URL")
	}

	if err == nil {
		t.Fatal("expected database error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}