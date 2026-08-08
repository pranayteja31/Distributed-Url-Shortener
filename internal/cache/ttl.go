package cache

import "time"

const MaxCacheTTL = time.Hour

func URLTTL(expiresAt *time.Time)time.Duration {
	//no expiry
	if expiresAt == nil {
		return 30 * time.Minute
	}
	//expiry is not null
	ttl := time.Until(*expiresAt)

	if ttl < 0 {
		return 0
	}
	if ttl > MaxCacheTTL {
		return MaxCacheTTL
	}
	return ttl

}