package db

import "strings"

// isDuplicateKeyError checks for Postgres SQLSTATE 23505
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505")
}
