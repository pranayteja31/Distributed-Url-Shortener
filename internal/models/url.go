package models

import "time"

type URL struct {
	ID          int64      `db:"id" json:"id"`
	ShortCode   string     `db:"short_code" json:"short_code"`
	OriginalURL string     `db:"original_url" json:"original_url"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	ClickCount  int64      `db:"click_count" json:"click_count"`
}