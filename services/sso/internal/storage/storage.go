package storage

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user is not found")
	ErrUserDontHasThisRole = errors.New("user don't have this role")
	ErrUserHasAnotherRole  = errors.New("user has an account with another role")
)
