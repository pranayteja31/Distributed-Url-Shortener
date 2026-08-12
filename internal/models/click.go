package models

import "time"

type URLClick struct {
	ID        int64     `db:"id" json:"id"`
	URLID     int64     `db:"url_id" json:"url_id"`
	ClickedAt time.Time `db:"clicked_at" json:"clicked_at"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	Referrer  string    `db:"referrer" json:"referrer"`
}
