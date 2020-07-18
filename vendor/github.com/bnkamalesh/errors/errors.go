// Package errors helps in wrapping errors with custom type as well as a user friendly message. This is particularly useful when responding to APIs
package errors

import (
	"net/http"
	"runtime"
	"strings"
)

type errType int

func (e errType) Int() int {
	return int(e)
}

// While adding a new Type, the respective helper functions should be added, also update the
// WriteHTTP method accordingly
const (
	// TypeInternal is error type for when there is an internal system error. e.g. Database errors
	TypeInternal errType = iota
	// TypeValidation is error type for when there is a validation error. e.g. invalid email address
	TypeValidation
	// TypeInputBody is error type for when an input data type error. e.g. invalid JSON
	TypeInputBody
	// TypeDuplicate is error type for when there's duplicate content
	TypeDuplicate
	// TypeUnauthenticated is error type when trying to access an authenticated API without authentication
	TypeUnauthenticated
	// TypeUnauthorized is error type for when there's an unauthorized access attempt
	TypeUnauthorized
	// TypeEmpty is error type for when an expected non-empty resource, is empty
	TypeEmpty
	// TypeNotFound is error type for an expected resource is not found e.g. user ID not found
	TypeNotFound
	// TypeMaximumAttempts is error type for attempting the same action more than allowed
	TypeMaximumAttempts
	// TypeSubscriptionExpired is error type for when a user's 'paid' account has expired
	TypeSubscriptionExpired
	// TypeDownstreamDependencyTimedout is error type for when a request to a downstream dependent service times out
	TypeDownstreamDependencyTimedout

	// DefaultMessage is the default user friendly message
	DefaultMessage = "unknown error occurred"
)

var (
	defaultErrType = TypeInternal
)

// Error is the struct which holds custom attributes
type Error struct {
	// original is the original error
	original error
	// Message is meant to be returned as response of API, so this should be a user-friendly message
	message string
	// Type is used to define the type of the error, e.g. Server error, validation error etc.
	eType    errType
	fileLine string
}

// Error is the implementation of error interface
func (e *Error) Error() string {
	if e.original != nil {
		// string concatenation with + is ~100x faster than fmt.Sprintf()
		return e.fileLine + " " + e.original.Error()
		/*
			// use the following code instead of the above return, to avoid nested filename:line
			err, _ := e.original.(*Error)
			if err != nil {
			    return err.Error()
			}
			return fmt.Sprintf("%s %s", e.fileLine, e.original.Error())
		*/
	}

	if e.message != "" {
		// string concatenation with + is ~100x faster than fmt.Sprintf()
		return e.fileLine + " " + e.message
	}

	// string concatenation with + is ~100x faster than fmt.Sprintf()
	return e.fileLine + " " + DefaultMessage
}

// Message returns the user friendly message stored in the error struct
func (e *Error) Message() string {
	messages := make([]string, 0, 5)
	messages = append(messages, e.message)

	err, _ := e.original.(*Error)
	for err != nil {
		messages = append(messages, err.message)
		err, _ = err.original.(*Error)
	}

	if len(messages) > 1 {
		return strings.Join(messages, ". ")
	}

	return e.message
}

// Unwrap implement's Go 1.13's Unwrap interface exposing the wrapped error
func (e *Error) Unwrap() error {
	if e.original == nil {
		return nil
	}

	return e.original
}

// Is implements the Is interface required by Go
func (e *Error) Is(err error) bool {
	o, _ := err.(*Error)
	if o == nil {
		return false
	}
	return o == e
}

// HTTPStatusCode is a convenience method used to get the appropriate HTTP response status code for the respective error type
func (e *Error) HTTPStatusCode() int {
	status := http.StatusInternalServerError
	switch e.eType {
	case TypeValidation:
		{
			status = http.StatusUnprocessableEntity
		}
	case TypeInputBody:
		{
			status = http.StatusBadRequest
		}

	case TypeDuplicate:
		{
			status = http.StatusConflict
		}

	case TypeUnauthenticated:
		{
			status = http.StatusUnauthorized
		}
	case TypeUnauthorized:
		{
			status = http.StatusForbidden
		}

	case TypeEmpty:
		{
			status = http.StatusGone
		}

	case TypeNotFound:
		{
			status = http.StatusNotFound

		}
	case TypeMaximumAttempts:
		{
			status = http.StatusTooManyRequests
		}
	case TypeSubscriptionExpired:
		{
			status = http.StatusPaymentRequired
		}
	}

	return status
}

// Type returns the error type as integer
func (e *Error) Type() errType {
	return e.eType
}

// New returns a new instance of Error with the relavant fields initialized
func New(msg string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, msg, file, line, defaultErrType)
}

// SetDefaultType will set the default error type, which is used in the 'New' function
func SetDefaultType(e errType) {
	defaultErrType = e
}
