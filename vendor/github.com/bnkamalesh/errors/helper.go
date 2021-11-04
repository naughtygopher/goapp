package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

func newerr(e error, message string, file string, line int, etype errType) *Error {
	return &Error{
		original: e,
		message:  message,
		eType:    etype,
		// '+' concatenation is ~100x faster than using fmt.Sprintf("%s:%d", file, line)
		fileLine: file + ":" + strconv.Itoa(line),
	}
}

func newerrf(e error, file string, line int, etype errType, format string, args ...interface{}) *Error {
	message := fmt.Sprintf(format, args...)
	return newerr(e, message, file, line, etype)
}

// Wrap is used to simply wrap an error with no custom error message with Error struct; with the error
// type being defaulted to `TypeInternal`
// If the error being wrapped is already of type Error, then its respective type is used
func Wrap(original error, msg ...string) *Error {
	message := strings.Join(msg, ". ")
	_, file, line, _ := runtime.Caller(1)
	e, _ := original.(*Error)
	if e == nil {
		return newerr(original, message, file, line, TypeInternal)
	}
	return newerr(original, message, file, line, e.eType)
}

func Wrapf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	e, _ := original.(*Error)
	if e == nil {
		return newerrf(original, file, line, TypeInternal, format, args...)
	}
	return newerrf(original, file, line, e.Type(), format, args...)
}

// WrapWithMsg [deprecated, use `Wrap`] wrap error with a user friendly message
func WrapWithMsg(original error, msg string) *Error {
	_, file, line, _ := runtime.Caller(1)
	e, _ := original.(*Error)
	if e == nil {
		return newerr(original, msg, file, line, TypeInternal)
	}
	return newerr(original, msg, file, line, e.eType)
}

// NewWithType returns an error instance with custom error type
func NewWithType(msg string, etype errType) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, msg, file, line, etype)
}

// NewWithTypef returns an error instance with custom error type. And formatted message
func NewWithTypef(etype errType, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, etype, format, args...)
}

// NewWithErrMsgType returns an error instance with custom error type and message
func NewWithErrMsgType(original error, message string, etype errType) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, etype)
}

// NewWithErrMsgTypef returns an error instance with custom error type and formatted message
func NewWithErrMsgTypef(original error, etype errType, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, etype, format, args...)
}

// Internal helper method for creating internal errors
func Internal(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeInternal)
}

// Internalf helper method for creating internal errors with formatted message
func Internalf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeInternal, format, args...)
}

// Validation is a helper function to create a new error of type TypeValidation
func Validation(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeValidation)
}

// Validationf is a helper function to create a new error of type TypeValidation, with formatted message
func Validationf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeValidation, format, args...)
}

// InputBody is a helper function to create a new error of type TypeInputBody
func InputBody(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeInputBody)
}

// InputBodyf is a helper function to create a new error of type TypeInputBody, with formatted message
func InputBodyf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeInputBody, format, args...)
}

// Duplicate is a helper function to create a new error of type TypeDuplicate
func Duplicate(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeDuplicate)
}

// Duplicatef is a helper function to create a new error of type TypeDuplicate, with formatted message
func Duplicatef(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeDuplicate, format, args...)
}

// Unauthenticated is a helper function to create a new error of type TypeUnauthenticated
func Unauthenticated(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeUnauthenticated)
}

// Unauthenticatedf is a helper function to create a new error of type TypeUnauthenticated, with formatted message
func Unauthenticatedf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeUnauthenticated, format, args...)

}

// Unauthorized is a helper function to create a new error of type TypeUnauthorized
func Unauthorized(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeUnauthorized)
}

// Unauthorizedf is a helper function to create a new error of type TypeUnauthorized, with formatted message
func Unauthorizedf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeUnauthorized, format, args...)
}

// Empty is a helper function to create a new error of type TypeEmpty
func Empty(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeEmpty)
}

// Emptyf is a helper function to create a new error of type TypeEmpty, with formatted message
func Emptyf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeEmpty, format, args...)
}

// NotFound is a helper function to create a new error of type TypeNotFound
func NotFound(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeNotFound)
}

// NotFoundf is a helper function to create a new error of type TypeNotFound, with formatted message
func NotFoundf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeNotFound, format, args...)
}

// MaximumAttempts is a helper function to create a new error of type TypeMaximumAttempts
func MaximumAttempts(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeMaximumAttempts)
}

// MaximumAttemptsf is a helper function to create a new error of type TypeMaximumAttempts, with formatted message
func MaximumAttemptsf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeMaximumAttempts, format, args...)
}

// SubscriptionExpired is a helper function to create a new error of type TypeSubscriptionExpired
func SubscriptionExpired(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeSubscriptionExpired)
}

// SubscriptionExpiredf is a helper function to create a new error of type TypeSubscriptionExpired, with formatted message
func SubscriptionExpiredf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeSubscriptionExpired, format, args...)
}

// DownstreamDependencyTimedout is a helper function to create a new error of type TypeDownstreamDependencyTimedout
func DownstreamDependencyTimedout(message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(nil, message, file, line, TypeDownstreamDependencyTimedout)
}

// DownstreamDependencyTimedoutf is a helper function to create a new error of type TypeDownstreamDependencyTimedout, with formatted message
func DownstreamDependencyTimedoutf(format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(nil, file, line, TypeDownstreamDependencyTimedout, format, args...)
}

// InternalErr helper method for creation internal errors which also accepts an original error
func InternalErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeInternal)
}

// InternalErr helper method for creation internal errors which also accepts an original error, with formatted message
func InternalErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeInternal, format, args...)
}

// ValidationErr helper method for creation validation errors which also accepts an original error
func ValidationErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeValidation)
}

// ValidationErr helper method for creation validation errors which also accepts an original error, with formatted message
func ValidationErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeValidation, format, args...)
}

// InputBodyErr is a helper function to create a new error of type TypeInputBody which also accepts an original error
func InputBodyErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeInputBody)
}

// InputBodyErrf is a helper function to create a new error of type TypeInputBody which also accepts an original error, with formatted message
func InputBodyErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeInputBody, format, args...)
}

// DuplicateErr is a helper function to create a new error of type TypeDuplicate which also accepts an original error
func DuplicateErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeDuplicate)
}

// DuplicateErrf is a helper function to create a new error of type TypeDuplicate which also accepts an original error, with formatted message
func DuplicateErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeDuplicate, format, args...)
}

// UnauthenticatedErr is a helper function to create a new error of type TypeUnauthenticated which also accepts an original error
func UnauthenticatedErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeUnauthenticated)
}

// UnauthenticatedErrf is a helper function to create a new error of type TypeUnauthenticated which also accepts an original error, with formatted message
func UnauthenticatedErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeUnauthenticated, format, args...)
}

// UnauthorizedErr is a helper function to create a new error of type TypeUnauthorized which also accepts an original error
func UnauthorizedErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeUnauthorized)
}

// UnauthorizedErrf is a helper function to create a new error of type TypeUnauthorized which also accepts an original error, with formatted message
func UnauthorizedErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeUnauthorized, format, args...)
}

// EmptyErr is a helper function to create a new error of type TypeEmpty which also accepts an original error
func EmptyErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeEmpty)
}

// EmptyErr is a helper function to create a new error of type TypeEmpty which also accepts an original error, with formatted message
func EmptyErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeEmpty, format, args...)
}

// NotFoundErr is a helper function to create a new error of type TypeNotFound which also accepts an original error
func NotFoundErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeNotFound)
}

// NotFoundErrf is a helper function to create a new error of type TypeNotFound which also accepts an original error, with formatted message
func NotFoundErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeNotFound, format, args...)
}

// MaximumAttemptsErr is a helper function to create a new error of type TypeMaximumAttempts which also accepts an original error
func MaximumAttemptsErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeMaximumAttempts)
}

// MaximumAttemptsErr is a helper function to create a new error of type TypeMaximumAttempts which also accepts an original error, with formatted message
func MaximumAttemptsErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeMaximumAttempts, format, args...)
}

// SubscriptionExpiredErr is a helper function to create a new error of type TypeSubscriptionExpired which also accepts an original error
func SubscriptionExpiredErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeSubscriptionExpired)
}

// SubscriptionExpiredErrf is a helper function to create a new error of type TypeSubscriptionExpired which also accepts an original error, with formatted message
func SubscriptionExpiredErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeSubscriptionExpired, format, args...)
}

// DownstreamDependencyTimedoutErr is a helper function to create a new error of type TypeDownstreamDependencyTimedout which also accepts an original error
func DownstreamDependencyTimedoutErr(original error, message string) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerr(original, message, file, line, TypeDownstreamDependencyTimedout)
}

// DownstreamDependencyTimedoutErrf is a helper function to create a new error of type TypeDownstreamDependencyTimedout which also accepts an original error, with formatted message
func DownstreamDependencyTimedoutErrf(original error, format string, args ...interface{}) *Error {
	_, file, line, _ := runtime.Caller(1)
	return newerrf(original, file, line, TypeDownstreamDependencyTimedout, format, args...)
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

// ErrWithoutTrace is a duplicate of Message, but with clearer name. The boolean is 'true' if the
// provided err is of type *Error
func ErrWithoutTrace(err error) (string, bool) {
	return Message(err)
}

// Message recursively concatenates all the messages set while creating/wrapping the errors. The boolean
// is 'true' if the provided error is of type *Err
func Message(err error) (string, bool) {
	derr, _ := err.(*Error)
	if derr != nil {
		return derr.Message(), true
	}
	return "", false
}

// HTTPStatusCode returns appropriate HTTP response status code based on type of the error. The boolean
// is 'true' if the provided error is of type *Err
func HTTPStatusCode(err error) (int, bool) {
	derr, _ := err.(*Error)
	if derr != nil {
		return derr.HTTPStatusCode(), true
	}
	return 0, false
}

// WriteHTTP is a convenience method which will check if the error is of type *Error and
// respond appropriately
func WriteHTTP(err error, w http.ResponseWriter) {
	// INFO: consider sending back "unknown server error" message
	status, msg, _ := HTTPStatusCodeMessage(err)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

// Type returns the errType if it's an instance of *Error, -1 otherwise
func Type(err error) errType {
	e, ok := err.(*Error)
	if !ok {
		return errType(-1)
	}
	return e.Type()
}

// Type returns the errType as integer if it's an instance of *Error, -1 otherwise
func TypeInt(err error) int {
	return Type(err).Int()
}

// HasType will check if the provided err type is available anywhere nested in the error
func HasType(err error, et errType) bool {
	if err == nil {
		return false
	}

	e, _ := err.(*Error)
	if e == nil {
		return HasType(errors.Unwrap(err), et)
	}

	if e.Type() == et {
		return true
	}

	return HasType(e.Unwrap(), et)
}
