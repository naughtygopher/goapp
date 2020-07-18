package errors

import (
	"net/http"
	"runtime"
	"strconv"
)

func newerr(e error, message string, file string, line int, etype errType) *Error {
	return &Error{
		original: e,
		message:  message,
		eType:    etype,
		// '+' concatenation is much faster than using fmt.Sprintf("%s:%d", file, line)
		fileLine: file + ":" + strconv.Itoa(line),
	}
}

// NewWithType returns an error instance with custom error type
func NewWithType(msg string, etype errType) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, msg, file, line, etype)
}

// NewWithErrMsgType returns an error instance with custom error type and message
func NewWithErrMsgType(e error, message string, etype errType) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(e, message, file, line, etype)
}

// Internal helper method for creation internal errors
func Internal(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeInternal)
}

// Validation is a helper function to create a new error of type TypeValidation
func Validation(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeValidation)
}

// InputBody is a helper function to create a new error of type TypeInputBody
func InputBody(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeInputBody)
}

// Duplicate is a helper function to create a new error of type TypeDuplicate
func Duplicate(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeDuplicate)
}

// Unauthenticated is a helper function to create a new error of type TypeUnauthenticated
func Unauthenticated(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeUnauthenticated)
}

// Unauthorized is a helper function to create a new error of type TypeUnauthorized
func Unauthorized(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeUnauthorized)
}

// Empty is a helper function to create a new error of type TypeEmpty
func Empty(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeEmpty)
}

// NotFound is a helper function to create a new error of type TypeNotFound
func NotFound(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeNotFound)
}

// MaximumAttempts is a helper function to create a new error of type TypeMaximumAttempts
func MaximumAttempts(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeMaximumAttempts)
}

// SubscriptionExpired is a helper function to create a new error of type TypeSubscriptionExpired
func SubscriptionExpired(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeSubscriptionExpired)
}

// DownstreamDependencyTimedout is a helper function to create a new error of type TypeDownstreamDependencyTimedout
func DownstreamDependencyTimedout(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeDownstreamDependencyTimedout)
}

// InternalErr helper method for creation internal errors which also accepts an original error
func InternalErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeInternal)
}

// ValidationErr helper method for creation validation errors which also accepts an original error
func ValidationErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeValidation)
}

// InputBodyErr is a helper function to create a new error of type TypeInputBody which also accepts an original error
func InputBodyErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeInputBody)
}

// DuplicateErr is a helper function to create a new error of type TypeDuplicate which also accepts an original error
func DuplicateErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeDuplicate)
}

// UnauthenticatedErr is a helper function to create a new error of type TypeUnauthenticated which also accepts an original error
func UnauthenticatedErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeUnauthenticated)
}

// UnauthorizedErr is a helper function to create a new error of type TypeUnauthorized which also accepts an original error
func UnauthorizedErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeUnauthorized)
}

// EmptyErr is a helper function to create a new error of type TypeEmpty which also accepts an original error
func EmptyErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeEmpty)
}

// NotFoundErr is a helper function to create a new error of type TypeNotFound which also accepts an original error
func NotFoundErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeNotFound)
}

// MaximumAttemptsErr is a helper function to create a new error of type TypeMaximumAttempts which also accepts an original error
func MaximumAttemptsErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeMaximumAttempts)
}

// SubscriptionExpiredErr is a helper function to create a new error of type TypeSubscriptionExpired which also accepts an original error
func SubscriptionExpiredErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeSubscriptionExpired)
}

// DownstreamDependencyTimedoutErr is a helper function to create a new error of type TypeDownstreamDependencyTimedout which also accepts an original error
func DownstreamDependencyTimedoutErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeDownstreamDependencyTimedout)
}

// HTTPStatusCodeMessage returns the appropriate HTTP status code, message, boolean for the error
// the boolean value is true if the error was of type *Error, false otherwise
func HTTPStatusCodeMessage(err error) (int, string, bool) {
	derr, _ := err.(*Error)
	if derr != nil {
		return derr.HTTPStatusCode(), derr.Message(), true
	}

	return http.StatusInternalServerError, err.Error(), false
}

// WriteHTTP is a convenience method which will check if the error is of type *Error and
// respond appropriately
func WriteHTTP(err error, w http.ResponseWriter) {
	// INFO: consider sending back "unknown server error" message
	status, msg, _ := HTTPStatusCodeMessage(err)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}
