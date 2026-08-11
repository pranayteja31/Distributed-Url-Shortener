package repository

import (
	"pranayteja31/Urlshortener/internal/models"

	"github.com/jmoiron/sqlx"
)

type ClickRepository struct {
	db *sqlx.DB
}

func NewClickRepository(db *sqlx.DB) *ClickRepository {
	return &ClickRepository{
		db: db,
	}
}

func (r *ClickRepository) Create(click *models.URLClick) error {
	query := `
		INSERT INTO url_clicks (
			url_id,
			clicked_at,
			ip_address,
			user_agent,
			referrer
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	return r.db.QueryRow(
		query,
		click.URLID,
		click.ClickedAt,
		click.IPAddress,
		click.UserAgent,
		click.Referrer,
	).Scan(&click.ID)
}