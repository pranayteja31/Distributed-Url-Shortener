package repository

import (
	"pranayteja31/Urlshortener/internal/models"

	"github.com/jmoiron/sqlx"
)

//storing the DB connection
type URLRepository struct {
	db *sqlx.DB
}
//constructor
func NewRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

//creating a new record
func (r *URLRepository)Create(url *models.URL) error{
	q := `INSERT INTO urls(
	short_code,
	original_code,
	created_at,
	expires_at,
	click_count
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING id;
	`
	//query operation
	err := r.db.QueryRow(q,url.ShortCode,url.OriginalURL,url.CreatedAt,url.ExpiresAt,url.ClickCount).Scan(&url.ID)
	return err
}
//finding the url by short code
func (r *URLRepository)FindByShortCode(code string) (*models.URL,error){
	var url models.URL
	q := `SELECT id,short_code,original_code,created_at,expires_at,click_count FROM urls WHERE short_code=$1; `
	err := r.db.Get(&url,q,code)
	if err != nil {
		return nil,err
	}
	return &url,nil
}

//finding the url by original code
func (r *URLRepository)FindByOriginalCode(code string) (*models.URL,error){
	var url models.URL
	q := `SELECT id,short_code,original_code,created_at,expires_at,click_count FROM urls WHERE original_code=$1; `
	err := r.db.Get(&url,q,code)
	if err != nil {
		return nil,err
	}
	return &url,nil
}

//finding by id
func (r *URLRepository)FindByID(id int64) (*models.URL,error) {
	var url models.URL
	q := `SELECT id,short_code,original_code,created_at,expires_at,click_count FROM urls WHERE id=$1; `
	err := r.db.Get(&url,q,id)
	if err != nil {
		return nil,err
	}
	return &url,nil
}

//update query
func (r *URLRepository)Update(url *models.URL) error {
	query := `
		UPDATE urls
		SET
			original_url = $1,
			expires_at = $2
		WHERE id = $3;
	`

	_, err := r.db.Exec(
		query,
		url.OriginalURL,
		url.ExpiresAt,
		url.ID,
	)

	return err
}

//deleting query
func (r *URLRepository)Delete(id int64) error {
	query := `
		DELETE FROM urls
		WHERE id = $1;
	`

	_, err := r.db.Exec(query, id)

	return err
}

//List all the values
func (r *URLRepository) List() ([]models.URL, error) {
	var urls []models.URL

	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count
		FROM urls
		ORDER BY created_at DESC;
	`

	err := r.db.Select(&urls, query)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

//increment the count of no of clicks
func (r *URLRepository)IncrementCount(id int64) error{
	q := `UPDATE urls SET click_count = click_count + 1 WHERE id=$1;`

	_,err := r.db.Exec(q,id)
	if err != nil {
		return err
	}
	return nil
}

