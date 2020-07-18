package errors

import "errors"

// Unwrap calls the Go builtin errors.UnUnwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is calls the Go builtin errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As calls the Go builtin errors.As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
