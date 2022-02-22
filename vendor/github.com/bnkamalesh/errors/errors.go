// Package errors helps in wrapping errors with custom type as well as a user friendly message. This is particularly useful when responding to APIs
package errors

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
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
	pcs      []uintptr
}

// Error is the implementation of error interface
func (e *Error) Error() string {
	if e.original != nil {
		// string concatenation with + is ~100x faster than fmt.Sprintf()
		return e.fileLine + ": " + e.message + "\n" + e.original.Error()
	}

	if e.message != "" {
		// string concatenation with + is ~100x faster than fmt.Sprintf()
		return e.fileLine + ": " + e.message
	}

	// string concatenation with + is ~100x faster than fmt.Sprintf()
	return e.fileLine + ": " + DefaultMessage
}

// ErrorWithoutFileLine prints the final string without the stack trace / file+line number
func (e *Error) ErrorWithoutFileLine() string {
	if e.original != nil {
		if e.message != "" {
			// string concatenation with + is ~100x faster than fmt.Sprintf()
			msg := e.message + ": "
			if o, ok := e.original.(*Error); ok {
				msg += o.ErrorWithoutFileLine()
			} else {
				msg += e.original.Error()
			}
			return msg
		}
		return e.original.Error()
	}

	if e.message != "" {
		return e.message
	}

	return e.fileLine
}

// Message returns the user friendly message stored in the error struct. It will ignore all errors
// which are not of type *Error
func (e *Error) Message() string {
	messages := make([]string, 0, 5)
	if e.message != "" {
		messages = append(messages, e.message)
	}

	err, _ := e.original.(*Error)
	for err != nil {
		if err.message == "" {
			err, _ = err.original.(*Error)
			continue
		}
		messages = append(messages, err.message)
		err, _ = err.original.(*Error)
	}

	if len(messages) > 0 {
		return strings.Join(messages, ": ")
	}

	return e.Error()
}

// Unwrap implement's Go 1.13's Unwrap interface exposing the wrapped error
func (e *Error) Unwrap() error {
	return e.original
}

// Is implements the Is interface required by Go
func (e *Error) Is(err error) bool {
	o, _ := err.(*Error)
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

// Format implements the verbs/directives supported by Error to be used in fmt annotated/formatted strings
/*
%v  - the same output as Message(). i.e. recursively get all the custom messages set by user
    - if any of the wrapped error is not of type *Error, that will *not* be displayed
%+v - recursively prints all the messages along with the file & line number. Also includes output of `Error()` of
non *Error types.

%s  - identical to %v
%+s - recursively prints all the messages without file & line number. Also includes output `Error()` of
non *Error types.
*/
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.Error())
		} else {
			_, _ = io.WriteString(s, e.Message())
		}
	case 's':
		{
			if s.Flag('+') {
				_, _ = io.WriteString(s, e.ErrorWithoutFileLine())
			} else {
				_, _ = io.WriteString(s, e.Message())
			}
		}
	}
}

func (e *Error) RuntimeFrames() *runtime.Frames {
	return runtime.CallersFrames(e.pcs)
}

func (e *Error) ProgramCounters() []uintptr {
	return e.pcs
}

func (e *Error) StackTrace() string {
	trace := make([]string, 0, 100)
	rframes := e.RuntimeFrames()
	frame, ok := rframes.Next()
	line := strconv.Itoa(frame.Line)
	trace = append(trace, frame.Function+"(): "+e.message)
	for ok {
		trace = append(trace, "\t"+frame.File+":"+line)
		frame, ok = rframes.Next()
	}
	return strings.Join(trace, "\n")
}

func (e *Error) StackTraceNoFormat() []string {
	trace := make([]string, 0, 100)
	rframes := e.RuntimeFrames()
	frame, ok := rframes.Next()
	line := strconv.Itoa(frame.Line)
	trace = append(trace, frame.Function+"(): "+e.message)
	for ok {
		trace = append(trace, frame.File+":"+line)
		frame, ok = rframes.Next()
	}
	return trace
}

// StackTraceCustomFormat lets you prepare a stacktrace in a custom format
/*
Supported directives:
%m - message
%p - file path
%l - line
%f - function
*/
func (e *Error) StackTraceCustomFormat(msgformat string, traceFormat string) string {
	rframes := e.RuntimeFrames()
	frame, ok := rframes.Next()

	message := strings.ReplaceAll(msgformat, "%m", e.message)
	message = strings.ReplaceAll(message, "%p", frame.File)
	message = strings.ReplaceAll(message, "%l", strconv.Itoa(frame.Line))
	message = strings.ReplaceAll(message, "%f", frame.Function)
	traces := make([]string, 0, 100)
	traces = append(traces, message)

	for ok {
		trace := strings.ReplaceAll(traceFormat, "%m", e.message)
		trace = strings.ReplaceAll(trace, "%p", frame.File)
		trace = strings.ReplaceAll(trace, "%l", strconv.Itoa(frame.Line))
		trace = strings.ReplaceAll(trace, "%f", frame.Function)

		traces = append(traces, trace)
		frame, ok = rframes.Next()
	}

	return strings.Join(traces, "")
}

// New returns a new instance of Error with the relavant fields initialized
func New(msg string) *Error {
	return newerr(nil, msg, defaultErrType)
}

func Newf(fromat string, args ...interface{}) *Error {
	return newerrf(nil, defaultErrType, fromat, args...)
}

// Errorf is a convenience method to create a new instance of Error with formatted message
// Important: %w directive is not supported, use fmt.Errorf if you're using the %w directive or
// use Wrap/Wrapf to wrap an error.
func Errorf(fromat string, args ...interface{}) *Error {
	return Newf(fromat, args...)
}

// SetDefaultType will set the default error type, which is used in the 'New' function
func SetDefaultType(e errType) {
	defaultErrType = e
}

// Stacktrace returns a string representation of the stacktrace, where each trace is separated by a newline and tab '\t'
func Stacktrace(err error) string {
	trace := make([]string, 0, 100)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			trace = append(trace, e.StackTrace())
		} else {
			trace = append(trace, err.Error())
		}
		err = Unwrap(err)
	}
	return strings.Join(trace, "\n")
}

// Stacktrace returns a string representation of the stacktrace, as a slice of string where each
// element represents the error message and traces.
func StacktraceNoFormat(err error) []string {
	trace := make([]string, 0, 100)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			trace = append(trace, e.StackTraceNoFormat()...)
		} else {
			trace = append(trace, err.Error())
		}
		err = Unwrap(err)
	}
	return trace
}

// StacktraceCustomFormat lets you prepare a stacktrace in a custom format
/*
msgformat - is used to format the line which prints message from Error.message
traceFormat - is used to format the line which prints trace
Supported directives:
%m - message if err type is *Error, otherwise output of `.Error()`
%p - file path, empty if type is not *Error
%l - line, empty if type is not *Error
%f - function, empty if type is not *Error
*/
func StacktraceCustomFormat(msgformat string, traceFormat string, err error) string {
	trace := make([]string, 0, 100)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			trace = append(trace, e.StackTraceCustomFormat(msgformat, traceFormat))
		} else {
			message := strings.ReplaceAll(msgformat, "%m", err.Error())
			message = strings.ReplaceAll(message, "%p", "")
			message = strings.ReplaceAll(message, "%l", "")
			message = strings.ReplaceAll(message, "%f", "")

			trace = append(
				trace,
				message,
			)
		}
		err = Unwrap(err)
	}
	return strings.Join(trace, "")
}

func ProgramCounters(err error) []uintptr {
	pcs := make([]uintptr, 0, 100)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			pcs = append(pcs, e.ProgramCounters()...)
		}
		err = Unwrap(err)
	}
	return pcs
}

func RuntimeFrames(err error) *runtime.Frames {
	pcs := ProgramCounters(err)
	return runtime.CallersFrames(pcs)
}
