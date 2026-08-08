package services

import (
	"context"
	"encoding/json"
	"errors"
	"pranayteja31/Urlshortener/internal/cache"
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
	"pranayteja31/Urlshortener/internal/utils"
	"time"
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
	// Generate a unique short code
	var shortCode string

	for {
		shortCode = utils.GenerateShortCode(6)

		existing, err := s.repo.FindByShortCode(shortCode)
		if err != nil {
			// If no rows found, code is available
			break
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
	return s.repo.FindByID(id)
}

//5.updation of the url
func (s *URLServices) UpdateURL(url *models.URL,exp int) (*models.URL, error) {
	existingUrl,err := s.repo.FindByID(url.ID)
	if err != nil {
		return nil,err
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
		return err
	}

	err = s.repo.Delete(id)
	if err != nil {
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
	if found && err == nil {
		var url models.URL
		if err:= json.Unmarshal(cacheData,&url); err != nil{
			if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt){
				return "", errors.New("URL has expired")
			}
			if err:= s.repo.IncrementCount(url.ID); err != nil {
				return "",err
			}
			
			return url.OriginalURL,nil
		}
	}
	// Fetch URL details
	urlDetails, err := s.repo.FindByShortCode(shortCode)
	if err != nil || urlDetails == nil {
		return "", err
	}

	// Check expiry (if expiry is set)
	if urlDetails.ExpiresAt != nil && time.Now().After(*urlDetails.ExpiresAt) {
		return "", errors.New("URL has expired")
	}

	// Increment click count
	if err := s.repo.IncrementCount(urlDetails.ID); err != nil {
		return "", err
	}

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