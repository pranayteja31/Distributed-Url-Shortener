package utils

import (
	"errors"
	"net/url"
	"strings"
)

const MaxURLLength = 2048

var (
	ErrEmptyURL           = errors.New("URL cannot be empty")
	ErrURLTooLong         = errors.New("URL exceeds maximum length")
	ErrInvalidURL         = errors.New("invalid URL")
	ErrUnsupportedScheme  = errors.New("unsupported URL scheme")
	ErrMissingURLHost     = errors.New("URL host is missing")
)

// ValidateURL validates a URL before it enters the business logic.
func ValidateURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return ErrEmptyURL
	}

	if len(rawURL) > MaxURLLength {
		return ErrURLTooLong
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Only HTTP and HTTPS URLs are allowed.
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrUnsupportedScheme
	}

	// A valid HTTP/HTTPS URL must contain a host.
	if parsedURL.Host == "" {
		return ErrMissingURLHost
	}

	return nil
}