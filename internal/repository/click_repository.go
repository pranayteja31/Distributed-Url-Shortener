package repository

import (
	"pranayteja31/Urlshortener/internal/models"

	"github.com/jmoiron/sqlx"
)

//repo struct
type ClickRepository struct {
	db *sqlx.DB
}
//constructor
func NewClickRepository(db *sqlx.DB) *ClickRepository {
	return &ClickRepository{
		db: db,
	}
}

//methods
func (r *ClickRepository)Create (click *models.URLClick) error {
	query := `INSERT INTO url_clicks(
		url_id,
		clicked_at,
		ip_address,
		user_agent,
		referrer
	)
	VALUES(&1,&2,&3,&4,&5)
	RETURNING id;
	`
	return r.db.QueryRow(query,click.ID,click.ClickedAt,click.IPAddress,click.UserAgent,click.Referrer).Scan(&click.ID)
}

