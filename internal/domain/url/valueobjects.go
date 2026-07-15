package url

import (
	"errors"
	"net/url"
	"regexp"
)

type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{1,62}[a-z0-9]$`)
var reservedSlugs = map[string]bool{
	"api": true, "admin": true, "login": true, "auth": true, "www": true,
}

func NewSlug(s string) (Slug, error) {
	if !slugRegex.MatchString(s) {
		return "", errors.New("invalid slug format: must be 3-64 chars, alphanumeric with hyphens/underscores")
	}
	if reservedSlugs[s] {
		return "", errors.New("invalid slug: reserved word")
	}
	return Slug(s), nil
}

type OriginalURL string

func NewOriginalURL(s string) (OriginalURL, error) {
	u, err := url.ParseRequestURI(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("invalid original url: must be a valid http/https URL")
	}
	return OriginalURL(s), nil
}
