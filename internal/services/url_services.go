package services

import (
	"errors"
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
	"pranayteja31/Urlshortener/internal/utils"
	"time"
)

//struct
type URLServices struct {
	repo *repository.URLRepository
}

//constructor
func NewURLServices(repo *repository.URLRepository) *URLServices {
	return &URLServices{
		repo: repo,
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
	return existingUrl,nil
}

//6. delete created url
func (s *URLServices) DeleteURL(id int64) error {
	return s.repo.Delete(id)
}

//7. list all the urls
func (s *URLServices) ListURLs() ([]models.URL, error) {
	return s.repo.List()
}

//8. increment the count when somebody clicks it
func (s *URLServices) RedirectURL(shortCode string) (string, error) {
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

	return urlDetails.OriginalURL, nil
}