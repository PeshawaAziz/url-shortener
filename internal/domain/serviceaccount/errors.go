package serviceaccount

import "errors"

var (
	ErrServiceAccountNotFound = errors.New("service account not found")
	ErrDuplicateName          = errors.New("service account name already exists in this tenant")
	ErrInactiveAccount        = errors.New("service account is inactive or deleted")
)
