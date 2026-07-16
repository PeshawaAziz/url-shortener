package url

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrSlugConflict     = errors.New("slug already exists")
	ErrInvalidSlug      = errors.New("invalid slug")
	ErrInvalidURL       = errors.New("invalid destination url")
	ErrMaxRetriesHit    = errors.New("max retries hit for slug generation")
	ErrLinkExpired      = errors.New("link has expired or cap reached")
	ErrRateLimited      = errors.New("has rate limited")
	ErrPasswordRequired = errors.New("password is required")
)
