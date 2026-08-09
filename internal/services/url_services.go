package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"pranayteja31/Urlshortener/internal/cache"
	"pranayteja31/Urlshortener/internal/metrics"
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
	"pranayteja31/Urlshortener/internal/utils"
	"time"
)

//custom errors
var (
	ErrURLNotFound = errors.New("url not found")
	ErrUrlExpired = errors.New("url has expired")
)

//struct
type URLServices struct {
	repo *repository.URLRepository
	cache cache.Cache
}

//constructor
func NewURLServices(repo *repository.URLRepository, cache cache.Cache) *URLServices {
	return &URLServices{
		repo: repo,
		cache: cache,
	}
}

//service of url
func (s *URLServices) CreateShortURL(originalURL string, exp int) (*models.URL,error) {
	//validate url
	if err:= utils.ValidateURL(originalURL); err != nil {
		return nil,err
	}
	// Generate a unique short code
	var shortCode string

	for {
		shortCode = utils.GenerateShortCode(6)

		existing, err := s.repo.FindByShortCode(shortCode)
		if err != nil {
			if errors.Is(err,sql.ErrNoRows) {
				break
			}
			return nil,err
		}

		if existing == nil {
			break
		}
	}

	now := time.Now()
	expiry := now.Add(time.Duration(exp) * 24 * time.Hour)

	newURL := models.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   now,
		ExpiresAt:   &expiry,
		ClickCount:  0,
	}
	err := s.repo.Create(&newURL)
    if err != nil {
        return nil, err
    }
	//redis cache of the new url created
	ctx := context.Background()

	data, err := json.Marshal(newURL)
	if err == nil {
		ttl := cache.URLTTL(newURL.ExpiresAt)
		if ttl > 0 {
			_ = s.cache.Set(
			ctx,
			cache.URLKey(newURL.ShortCode),
			data,
			ttl,
			)
		}
		
	}
	return &newURL,nil
}
//get url
func (s *URLServices) GetURL(id int64) (*models.URL,error) {
	url,err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return nil,ErrURLNotFound
		}
		return nil,err
	}
	if url == nil {
		return nil, ErrURLNotFound
	}
	return url,nil
}

//5.updation of the url
func (s *URLServices) UpdateURL(url *models.URL,exp int) (*models.URL, error) {
	//validate url
	if err:= utils.ValidateURL(url.OriginalURL); err != nil {
		return nil,err
	}
	//
	existingUrl,err := s.repo.FindByID(url.ID)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return nil,ErrURLNotFound
		}
		return nil,err
	}
	if existingUrl == nil {
		return nil, ErrURLNotFound
	}
	existingUrl.OriginalURL = url.OriginalURL
	
	expiresAt := time.Now().Add(time.Duration(exp) * 24 * time.Hour)
	existingUrl.ExpiresAt = &expiresAt

	
	err = s.repo.Update(existingUrl)
	if err != nil{
		return nil,err
	}
	ctx := context.Background()

	_ = s.cache.Delete(ctx,cache.URLKey(existingUrl.ShortCode))
	return existingUrl,nil
}

//6. delete created url
func (s *URLServices) DeleteURL(id int64) error {

	url, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return ErrURLNotFound
		}
		return err
	}
	if url == nil {
		return ErrURLNotFound
	}
	err = s.repo.Delete(id)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return ErrURLNotFound
		}
		return err
	}

	ctx := context.Background()

	_ = s.cache.Delete(
		ctx,
		cache.URLKey(url.ShortCode),
	)

	return nil
}
//7. list all the urls
func (s *URLServices) ListURLs() ([]models.URL, error) {
	return s.repo.List()
}

//8. increment the count when somebody clicks it
func (s *URLServices) RedirectURL(shortCode string) (string, error) {
	//redis cache logic
	ctx := context.Background()
	key := cache.URLKey(shortCode)

	cacheData,found, err := s.cache.Get(ctx,key)
	if err == nil && found {
		metrics.CacheHits.Inc()
		var url models.URL
		if err:= json.Unmarshal(cacheData,&url); err == nil{

			fmt.Println("CACHE HIT → Redis:", shortCode)

			if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt){
				return "", ErrUrlExpired
			}
			if err:= s.repo.IncrementCount(url.ID); err != nil {
				return "",err
			}
			
			return url.OriginalURL,nil
		}
	}
	metrics.CacheMisses.Inc()
	fmt.Println("CACHE MISS → Redis:", shortCode)

	// Fetch URL details
	postgresStart := time.Now()
	urlDetails, err := s.repo.FindByShortCode(shortCode)
	metrics.ObservePostgresLatency(postgresStart)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}

		return "", err
	}

	if urlDetails == nil {
		return "", ErrURLNotFound
	}

	// Check expiry (if expiry is set)
	if urlDetails.ExpiresAt != nil && time.Now().After(*urlDetails.ExpiresAt) {
		return "", ErrUrlExpired
	}

	// Increment click count
	if err := s.repo.IncrementCount(urlDetails.ID); err != nil {
		return "", err
	}
	//redis cache set
	data, err := json.Marshal(urlDetails)
	if err == nil {
		ttl := cache.URLTTL(urlDetails.ExpiresAt)
		if ttl > 0 {
			_ = s.cache.Set(ctx,key,data,ttl)
		}
	}
	//return from postgreSql server
	return urlDetails.OriginalURL, nil
}