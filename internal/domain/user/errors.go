package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyTaken  = errors.New("email already taken for this tenant")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrEmailNotVerified   = errors.New("email not verified")
)
