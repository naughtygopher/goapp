// Package errors helps in wrapping errors with custom type as well as a user friendly message. This is particularly useful when responding to APIs
package errors

import (
	"bytes"
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
	eType errType
	pcs   []uintptr
	pc    uintptr
}

func (e *Error) fileLine() string {
	if e.pc == 0 {
		return ""
	}

	frames := runtime.CallersFrames([]uintptr{e.pc + 1})
	frame, _ := frames.Next()

	buff := bytes.NewBuffer(make([]byte, 0, 128))
	buff.WriteString(frame.File)
	buff.WriteString(":")
	buff.WriteString(strconv.Itoa(frame.Line))

	return buff.String()
}

// Error is the implementation of error interface
func (e *Error) Error() string {
	str := bytes.NewBuffer(make([]byte, 0, 128))
	str.WriteString(e.fileLine())
	if str.Len() != 0 {
		str.WriteString(": ")
	}

	if e.original != nil {
		str.WriteString(e.message)
		str.WriteString("\n")
		str.WriteString(e.original.Error())
		return str.String()
	}

	if e.message != "" {
		str.WriteString(e.message)
		return str.String()
	}

	str.WriteString(DefaultMessage)

	return str.String()
}

// ErrorWithoutFileLine prints the final string without the stack trace / file+line number
func (e *Error) ErrorWithoutFileLine() string {
	if e.original != nil {
		if e.message != "" {
			msg := bytes.NewBuffer(make([]byte, 0, 128))
			msg.WriteString(e.message)
			msg.WriteString(": ")
			if o, ok := e.original.(*Error); ok {
				msg.WriteString(o.ErrorWithoutFileLine())
			} else {
				msg.WriteString(e.original.Error())
			}
			return msg.String()
		}
		return e.original.Error()
	}

	if e.message != "" {
		return e.message
	}

	return e.fileLine()
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
	return runtime.CallersFrames(e.ProgramCounters())
}

func (e *Error) ProgramCounters() []uintptr {
	return e.pcs
}

func (e *Error) StackTrace() []string {
	rframes := e.RuntimeFrames()
	frame, ok := rframes.Next()
	buff := bytes.NewBuffer(make([]byte, 0, 128))
	buff.WriteString(frame.Function)
	buff.WriteString("(): ")
	buff.WriteString(e.message)

	trace := make([]string, 0, len(e.ProgramCounters()))
	trace = append(trace, buff.String())
	for ok {
		buff.Reset()
		buff.WriteString("\t")
		buff.WriteString(frame.File)
		buff.WriteString(":")
		buff.WriteString(strconv.Itoa(frame.Line))
		trace = append(trace, buff.String())
		frame, ok = rframes.Next()
	}
	return trace
}

func (e *Error) StackTraceNoFormat() []string {
	rframes := e.RuntimeFrames()
	frame, ok := rframes.Next()
	line := strconv.Itoa(frame.Line)

	buff := bytes.NewBuffer(make([]byte, 0, 128))
	buff.WriteString(frame.Function)
	buff.WriteString("(): ")
	buff.WriteString(e.message)

	trace := make([]string, 0, len(e.ProgramCounters()))
	trace = append(trace, buff.String())
	for ok {
		buff.Reset()
		buff.WriteString(frame.File)
		buff.WriteString(":")
		buff.WriteString(line)
		trace = append(trace, buff.String())
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
func (e *Error) StackTraceCustomFormat(msgformat string, traceFormat string) []string {
	rframes := e.RuntimeFrames()
	frame, ok := rframes.Next()

	message := strings.ReplaceAll(msgformat, "%m", e.message)
	message = strings.ReplaceAll(message, "%p", frame.File)
	message = strings.ReplaceAll(message, "%l", strconv.Itoa(frame.Line))
	message = strings.ReplaceAll(message, "%f", frame.Function)
	traces := make([]string, 0, len(e.ProgramCounters()))
	traces = append(traces, message)

	for ok {
		trace := strings.ReplaceAll(traceFormat, "%m", e.message)
		trace = strings.ReplaceAll(trace, "%p", frame.File)
		trace = strings.ReplaceAll(trace, "%l", strconv.Itoa(frame.Line))
		trace = strings.ReplaceAll(trace, "%f", frame.Function)

		traces = append(traces, trace)
		frame, ok = rframes.Next()
	}

	return traces
}

// New returns a new instance of Error with the relavant fields initialized
func New(msg string) *Error {
	return newerr(nil, msg, defaultErrType, 3)
}

func Newf(fromat string, args ...interface{}) *Error {
	return newerrf(nil, defaultErrType, 4, fromat, args...)
}

// Errorf is a convenience method to create a new instance of Error with formatted message
// Important: %w directive is not supported, use fmt.Errorf if you're using the %w directive or
// use Wrap/Wrapf to wrap an error.
func Errorf(fromat string, args ...interface{}) *Error {
	return newerrf(nil, defaultErrType, 4, fromat, args...)
}

// SetDefaultType will set the default error type, which is used in the 'New' function
func SetDefaultType(e errType) {
	defaultErrType = e
}

// Stacktrace returns a string representation of the stacktrace, where each trace is separated by a newline and tab '\t'
func Stacktrace(err error) string {
	trace := make([][]string, 0, 128)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			trace = append(trace, e.StackTrace())
		} else {
			trace = append(trace, []string{err.Error()})
		}
		err = Unwrap(err)
	}

	lookup := map[string]struct{}{}
	for idx := len(trace) - 1; idx >= 0; idx-- {
		list := trace[idx]
		uniqueList := make([]string, 0, len(list))
		for _, line := range list {
			_, ok := lookup[line]
			if ok {
				break
			}
			uniqueList = append(uniqueList, line)
			lookup[line] = struct{}{}
		}
		trace[idx] = uniqueList
	}
	final := make([]string, 0, len(trace)*3)
	for _, list := range trace {
		final = append(final, list...)
	}

	return strings.Join(final, "\n")
}

// Stacktrace returns a string representation of the stacktrace, as a slice of string where each
// element represents the error message and traces.
func StacktraceNoFormat(err error) []string {
	trace := make([][]string, 0, 128)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			trace = append(trace, e.StackTraceNoFormat())
		} else {
			trace = append(trace, []string{err.Error()})
		}
		err = Unwrap(err)
	}

	lookup := map[string]struct{}{}
	for idx := len(trace) - 1; idx >= 0; idx-- {
		list := trace[idx]
		uniqueList := make([]string, 0, len(list))
		for _, line := range list {
			_, ok := lookup[line]
			if ok {
				break
			}
			uniqueList = append(uniqueList, line)
			lookup[line] = struct{}{}
		}
		trace[idx] = uniqueList
	}
	final := make([]string, 0, len(trace)*3)
	for _, list := range trace {
		final = append(final, list...)
	}

	return final
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
	trace := make([][]string, 0, 128)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			trace = append(trace, e.StackTraceCustomFormat(msgformat, traceFormat))
		} else {
			message := strings.ReplaceAll(msgformat, "%m", err.Error())
			message = strings.ReplaceAll(message, "%p", "")
			message = strings.ReplaceAll(message, "%l", "")
			message = strings.ReplaceAll(message, "%f", "")
			trace = append(trace, []string{message})
		}
		err = Unwrap(err)
	}

	lookup := map[string]struct{}{}
	for idx := len(trace) - 1; idx >= 0; idx-- {
		list := trace[idx]
		uniqueList := make([]string, 0, len(list))
		for _, line := range list {
			_, ok := lookup[line]
			if ok {
				break
			}
			uniqueList = append(uniqueList, line)
			lookup[line] = struct{}{}
		}
		trace[idx] = uniqueList
	}

	final := make([]string, 0, len(trace)*3)
	for _, list := range trace {
		final = append(final, list...)
	}

	return strings.Join(final, "")
}

func ProgramCounters(err error) []uintptr {
	pcs := make([][]uintptr, 0, 128)
	for err != nil {
		e, ok := err.(*Error)
		if ok {
			pcs = append(pcs, e.ProgramCounters())
		}
		err = Unwrap(err)
	}

	lookup := map[uintptr]struct{}{}
	for idx := len(pcs) - 1; idx >= 0; idx-- {
		list := pcs[idx]
		uniqueList := make([]uintptr, 0, len(list))
		for _, line := range list {
			_, ok := lookup[line]
			if ok {
				break
			}
			uniqueList = append(uniqueList, line)
			lookup[line] = struct{}{}
		}
		pcs[idx] = uniqueList
	}
	final := make([]uintptr, 0, len(pcs)*3)
	for _, list := range pcs {
		final = append(final, list...)
	}
	return final
}

func RuntimeFrames(err error) *runtime.Frames {
	pcs := ProgramCounters(err)
	return runtime.CallersFrames(pcs)
}

func StacktraceFromPcs(err error) string {
	pcs := ProgramCounters(err)
	frames := runtime.CallersFrames(pcs)
	frame, hasMore := frames.Next()
	lines := make([]string, 0, len(pcs))
	if !hasMore {
		lines = append(lines, frame.File+":"+strconv.Itoa(frame.Line)+":"+frame.Function+"()")
	}
	for hasMore {
		lines = append(lines, frame.File+":"+strconv.Itoa(frame.Line)+":"+frame.Function+"()")
		frame, hasMore = frames.Next()
	}

	return strings.Join(lines, "\n")
}
