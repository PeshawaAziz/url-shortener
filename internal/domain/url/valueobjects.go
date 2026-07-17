package url

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{1,62}[a-z0-9]$`)

var reservedSlugs = map[string]bool{
	"api": true, "admin": true, "login": true, "auth": true, "www": true,
}

var homoglyphs = map[string]string{
	"0": "o",
	"1": "l",
	"l": "1",
	"O": "o",
}

func NewSlug(ctx context.Context, s string, checker ReservedSlugChecker) (Slug, error) {
	if s == "" {
		return "", errors.New("slug cannot be empty")
	}

	normalized := strings.ToLower(s)

	for lookalike, replacement := range homoglyphs {
		normalized = strings.ReplaceAll(normalized, lookalike, replacement)
	}

	if !slugRegex.MatchString(normalized) {
		return "", errors.New("invalid slug format: must be 3-64 chars, alphanumeric with hyphens/underscores")
	}

	if checker.IsReserved(ctx, normalized) {
		return "", errors.New("invalid slug: reserved word")
	}

	return Slug(normalized), nil
}

type OriginalURL string

func NewOriginalURL(s string) (OriginalURL, error) {
	u, err := url.ParseRequestURI(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("invalid original url: must be a valid http/https URL")
	}
	return OriginalURL(s), nil
}
