package services

import (
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
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
//1. create a short url
func (s* URLServices) CreateShortURL(url *models.URL) error{
	return s.repo.Create(url)
}

//2. get url by given id
func (s *URLServices)GetURLById(id int64)(*models.URL,error){
	return s.repo.FindByID(id)
}
//3. get original url
func (s *URLServices) GetURLByShortCode(code string) (*models.URL, error) {
	return s.repo.FindByShortCode(code)
}
//4. get the short url of the org url
func (s *URLServices) GetURLByOriginalURL(originalURL string) (*models.URL, error) {
	return s.repo.FindByOriginalCode(originalURL)
}
//5.updation of the url
func (s *URLServices) UpdateURL(url *models.URL) error {
	return s.repo.Update(url)
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
func (s *URLServices) RedirectURL(id int64) error {
	return s.repo.IncrementCount(id)
}