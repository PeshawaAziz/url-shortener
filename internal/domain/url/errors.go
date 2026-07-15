package url

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrSlugConflict     = errors.New("slug already exists for this tenant")
	ErrClickCapExceeded = errors.New("click cap exceeded")
	ErrURLExpired       = errors.New("url has expired")
	ErrURLInactive      = errors.New("url is inactive")
)
