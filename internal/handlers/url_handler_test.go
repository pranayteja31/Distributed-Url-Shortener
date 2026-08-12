package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/services"

	"github.com/gin-gonic/gin"
)

type mockURLService struct {
	createShortURLFunc func(string, int) (*models.URL, error)
	redirectURLFunc    func(string, string, string, string) (string, error)
	getURLFunc         func(int64) (*models.URL, error)
	listURLsFunc       func() ([]models.URL, error)
	updateURLFunc      func(*models.URL, int) (*models.URL, error)
	deleteURLFunc      func(int64) error
}

func (m *mockURLService) CreateShortURL(
	url string,
	exp int,
) (*models.URL, error) {
	return m.createShortURLFunc(url, exp)
}

func (m *mockURLService) RedirectURL(
	shortCode string,
	ip string,
	userAgent string,
	referer string,
) (string, error) {
	return m.redirectURLFunc(
		shortCode,
		ip,
		userAgent,
		referer,
	)
}

func (m *mockURLService) GetURL(id int64) (*models.URL, error) {
	return m.getURLFunc(id)
}

func (m *mockURLService) ListURLs() ([]models.URL, error) {
	return m.listURLsFunc()
}

func (m *mockURLService) UpdateURL(
	url *models.URL,
	exp int,
) (*models.URL, error) {
	return m.updateURLFunc(url, exp)
}

func (m *mockURLService) DeleteURL(id int64) error {
	return m.deleteURLFunc(id)
}
func TestCreateShortURLHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		createShortURLFunc: func(url string, exp int) (*models.URL, error) {
			return &models.URL{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: url,
			}, nil
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.POST("/urls", handler.CreateShortUrl)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader("orgUrl=https://example.com&exp=30"),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}
}
func TestCreateShortURLHandler_MissingURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{}

	handler := NewHandler(mockService)

	router := gin.New()
	router.POST("/urls", handler.CreateShortUrl)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader("exp=30"),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
func TestCreateShortURLHandler_InvalidExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{}

	handler := NewHandler(mockService)

	router := gin.New()
	router.POST("/urls", handler.CreateShortUrl)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader("orgUrl=https://example.com&exp=-10"),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
func TestGetURLHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		getURLFunc: func(id int64) (*models.URL, error) {
			return &models.URL{
				ID:          id,
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
			}, nil
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/urls/:id", handler.GetUrl)

	req := httptest.NewRequest(
		http.MethodGet,
		"/urls/1",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
func TestGetURLHandler_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/urls/:id", handler.GetUrl)

	req := httptest.NewRequest(
		http.MethodGet,
		"/urls/abc",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
func TestGetURLHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		getURLFunc: func(id int64) (*models.URL, error) {
			return nil, services.ErrURLNotFound
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/urls/:id", handler.GetUrl)

	req := httptest.NewRequest(
		http.MethodGet,
		"/urls/999",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
func TestRedirectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		redirectURLFunc: func(
			shortCode string,
			ip string,
			userAgent string,
			referer string,
		) (string, error) {
			return "https://example.com", nil
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/:shortCode", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		nil,
	)

	req.Header.Set("User-Agent", "test-agent")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTemporaryRedirect,
			rec.Code,
		)
	}

	if location := rec.Header().Get("Location"); location != "https://example.com" {
		t.Errorf(
			"expected Location https://example.com, got %s",
			location,
		)
	}
}
func TestRedirectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		redirectURLFunc: func(
			shortCode string,
			ip string,
			userAgent string,
			referer string,
		) (string, error) {
			return "", services.ErrURLNotFound
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/:shortCode", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/missing",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
func TestRedirectHandler_Expired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		redirectURLFunc: func(
			shortCode string,
			ip string,
			userAgent string,
			referer string,
		) (string, error) {
			return "", services.ErrURLExpired
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/:shortCode", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/expired",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusGone {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusGone,
			rec.Code,
		)
	}
}
func TestListURLsHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		listURLsFunc: func() ([]models.URL, error) {
			return []models.URL{
				{
					ID:          1,
					ShortCode:   "abc123",
					OriginalURL: "https://example.com",
				},
				{
					ID:          2,
					ShortCode:   "xyz789",
					OriginalURL: "https://google.com",
				},
			}, nil
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.GET("/urls", handler.ListUrls)

	req := httptest.NewRequest(
		http.MethodGet,
		"/urls",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
func TestDeleteURLHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		deleteURLFunc: func(id int64) error {
			return nil
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.DELETE("/urls/:id", handler.DeleteUrl)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/urls/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
func TestDeleteURLHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		deleteURLFunc: func(id int64) error {
			return services.ErrURLNotFound
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.DELETE("/urls/:id", handler.DeleteUrl)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/urls/999",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
func TestUpdateURLHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{
		updateURLFunc: func(
			url *models.URL,
			exp int,
		) (*models.URL, error) {
			url.ShortCode = "abc123"
			return url, nil
		},
	}

	handler := NewHandler(mockService)

	router := gin.New()
	router.PUT("/urls/:id", handler.UpdateUrl)

	req := httptest.NewRequest(
		http.MethodPut,
		"/urls/10",
		strings.NewReader(
			"orgUrl=https://example.com&exp=30",
		),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
func TestUpdateURLHandler_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{}

	handler := NewHandler(mockService)

	router := gin.New()
	router.PUT("/urls/:id", handler.UpdateUrl)

	req := httptest.NewRequest(
		http.MethodPut,
		"/urls/abc",
		strings.NewReader(
			"orgUrl=https://example.com&exp=30",
		),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
func TestUpdateURLHandler_MissingURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockURLService{}

	handler := NewHandler(mockService)

	router := gin.New()
	router.PUT("/urls/:id", handler.UpdateUrl)

	req := httptest.NewRequest(
		http.MethodPut,
		"/urls/10",
		strings.NewReader("exp=30"),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
