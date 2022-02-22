<p align="center"><img src="https://user-images.githubusercontent.com/1092882/87815217-d864a680-c882-11ea-9c94-24b67f7125fe.png" alt="errors gopher" width="256px"/></p>

[![Build Status](https://travis-ci.org/bnkamalesh/errors.svg?branch=master)](https://travis-ci.org/bnkamalesh/errors)
[![codecov](https://codecov.io/gh/bnkamalesh/errors/branch/master/graph/badge.svg)](https://codecov.io/gh/bnkamalesh/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/bnkamalesh/errors)](https://goreportcard.com/report/github.com/bnkamalesh/errors)
[![Maintainability](https://api.codeclimate.com/v1/badges/a86629ab167695d4db7f/maintainability)](https://codeclimate.com/github/bnkamalesh/errors)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](https://pkg.go.dev/github.com/bnkamalesh/errors?tab=doc)
[![](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#error-handling)

# Errors v0.9.0

Errors package is a drop-in replacement of the built-in Go errors package with no external dependencies. It lets you create errors of 11 different types which should handle most of the use cases. Some of them are a bit too specific for web applications, but useful nonetheless. Following are the primary features of this package:

1. Multiple (11) error types
2. User friendly message
3. Stacktrace - formatted, unfromatted, custom format (refer tests in errors_test.go)
4. Retrieve the Program Counters, for compatibility external libraries which generate their own stacktrace
5. Retrieve *runtime.Frames using `errors.RuntimeFrames(err error)`, for compatibility external libraries which generate their own stacktrace
6. HTTP status code and user friendly message (wrapped messages are concatenated) for all error types
7. Helper functions to generate each error type
8. Helper function to get error Type, error type as int, check if error type is wrapped anywhere in chain
9. fmt.Formatter support

In case of nested errors, the messages & errors are also looped through the full chain of errors.

### Prerequisites

Go 1.13+

### Available error types

1. TypeInternal - is the error type for when there is an internal system error. e.g. Database errors
2. TypeValidation - is the error type for when there is a validation error. e.g. invalid email address
3. TypeInputBody - is the error type for when the input data is invalid. e.g. invalid JSON
4. TypeDuplicate - is the error type for when there's duplicate content. e.g. user with email already exists (when trying to register a new user)
5. TypeUnauthenticated - is the error type when trying to access an authenticated API without authentication
6. TypeUnauthorized - is the error type for when there's an unauthorized access attempt
7. TypeEmpty - is the error type for when an expected non-empty resource, is empty
8. TypeNotFound - is the error type for an expected resource is not found. e.g. user ID not found
9. TypeMaximumAttempts - is the error type for attempting the same action more than an allowed threshold
10. TypeSubscriptionExpired - is the error type for when a user's 'paid' account has expired
11. TypeDownstreamDependencyTimedout - is the error type for when a request to a downstream dependent service times out

Helper functions are available for all the error types. Each of them have 2 helper functions, one which accepts only a string, and the other which accepts an original error as well as a user friendly message.

All the dedicated error type functions are documented [here](https://pkg.go.dev/github.com/bnkamalesh/errors?tab=doc#DownstreamDependencyTimedout). Names are consistent with the error type, e.g. errors.Internal(string) and errors.InternalErr(error, string)

### User friendly messages

More often than not, when writing APIs, we'd want to respond with an easier to undersand user friendly message. Instead of returning the raw error. And log the raw error.

There are helper functions for all the error types, when in need of setting a friendly message, there are helper functions have a suffix 'Err'. All such helper functions accept the original error and a string.

```golang
package main

import (
	"fmt"

	"github.com/bnkamalesh/errors"
)

func Bar() error {
	return fmt.Errorf("hello %s", "world!")
}

func Foo() error {
	err := Bar()
	if err != nil {
		return errors.InternalErr(err, "bar is not happy")
	}
	return nil
}

func main() {
	err := Foo()
	
	fmt.Println("err:", err)
	fmt.Println("\nerr.Error():", err.Error())

	fmt.Printf("\nformatted +v: %+v\n", err)
	fmt.Printf("\nformatted v: %v\n", err)
	fmt.Printf("\nformatted +s: %+s\n", err)
	fmt.Printf("\nformatted s: %s\n", err)

	_, msg, _ := errors.HTTPStatusCodeMessage(err)
	fmt.Println("\nmsg:", msg)
}
```

Output 
```
err: bar is not happy

err.Error(): /path/to/file.go:16: bar is not happy
hello world!

formatted +v: /path/to/file.go:16: bar is not happy
hello world!

formatted v: bar is not happy

formatted +s: bar is not happy: hello world!

formatted s: bar is not happy

msg: bar is not happy
```

[Playground link](https://go.dev/play/p/-WzDH46f_U5)

### File & line number prefixed to errors

A common annoyance with Go errors which most people are aware of is, figuring out the origin of the error, especially when there are nested function calls. Ever since error annotation was introduced in Go, a lot of people have tried using it to trace out an errors origin by giving function names, contextual message etc in it. e.g. `fmt.Errorf("database query returned error %w", err)`. However this errors package, whenever you call the Go error interface's `Error() string` function, prints the error prefixed by the filepath and line number. It'd look like `../Users/JohnDoe/apps/main.go:50 hello world` where 'hello world' is the error message.

### HTTP status code & message

The function `errors.HTTPStatusCodeMessage(error) (int, string, bool)` returns the HTTP status code, message, and a boolean value. The boolean is true, if the error is of type *Error from this package. If error is nested, it unwraps and returns a single concatenated message. Sample described in the 'How to use?' section

## How to use?

A sample was already shown in the user friendly message section, following one would show a few more scenarios.

```golang
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/webgo/v4"
	"github.com/bnkamalesh/webgo/v4/middleware"
)

func bar() error {
	return fmt.Errorf("%s %s", "sinking", "bar")
}

func bar2() error {
	err := bar()
	if err != nil {
		return errors.InternalErr(err, "bar2 was deceived by bar1 :(")
	}
	return nil
}

func foo() error {
	err := bar2()
	if err != nil {
		return errors.InternalErr(err, "we lost bar2!")
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := foo()
	if err != nil {
		// log the error on your server for troubleshooting
		fmt.Println(err.Error())
		// respond to request with friendly msg
		status, msg, _ := errors.HTTPStatusCodeMessage(err)
		webgo.SendError(w, msg, status)
		return
	}

	webgo.R200(w, "yay!")
}

func routes() []*webgo.Route {
	return []*webgo.Route{
		&webgo.Route{
			Name:    "home",
			Method:  http.MethodGet,
			Pattern: "/",
			Handlers: []http.HandlerFunc{
				handler,
			},
		},
	}
}

func main() {
	router := webgo.NewRouter(&webgo.Config{
		Host:         "",
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}, routes())

	router.UseOnSpecialHandlers(middleware.AccessLog)
	router.Use(middleware.AccessLog)
	router.Start()
}

```

[webgo](https://github.com/bnkamalesh/webgo) was used to illustrate the usage of the function, `errors.HTTPStatusCodeMessage`. It returns the appropriate http status code, user friendly message stored within, and a boolean value. Boolean value is `true` if the returned error of type *Error.
Since we get the status code and message separately, when using any web framework, you can set values according to the respective framework's native functions. In case of Webgo, it wraps errors in a struct of its own. Otherwise, you could directly respond to the HTTP request by calling `errors.WriteHTTP(error,http.ResponseWriter)`. 

Once the app is running, you can check the response by opening `http://localhost:8080` on your browser. Or on terminal
```bash
$ curl http://localhost:8080
{"errors":"we lost bar2!. bar2 was deceived by bar1 :(","status":500} // output
```

And the `fmt.Println(err.Error())` generated output on stdout would be:
```bash
/Users/username/go/src/errorscheck/main.go:28 /Users/username/go/src/errorscheck/main.go:20 sinking bar
```

## Benchmark [2021-12-13]

```bash
$ go version
go version go1.17.4 linux/amd64

$ go test -benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/bnkamalesh/errors
cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
Benchmark_Internal-8                             772088       1412 ns/op    1272 B/op    5 allocs/op
Benchmark_Internalf-8                            695674       1692 ns/op    1296 B/op    6 allocs/op
Benchmark_InternalErr-8                          822500       1404 ns/op    1272 B/op    5 allocs/op
Benchmark_InternalGetError-8                     881791       1319 ns/op    1368 B/op    6 allocs/op
Benchmark_InternalGetErrorWithNestedError-8      712803       1488 ns/op    1384 B/op    6 allocs/op
Benchmark_InternalGetMessage-8                   927864       1237 ns/op    1272 B/op    5 allocs/op
Benchmark_InternalGetMessageWithNestedError-8    761164       1675 ns/op    1296 B/op    6 allocs/op
Benchmark_HTTPStatusCodeMessage-8                29116684     41.62 ns/op   16 B/op      1 allocs/op
BenchmarkHasType-8                               100000000    11.50 ns/op   0 B/op       0 allocs/op
PASS
ok  	github.com/bnkamalesh/errors	10.604s
```

## Contributing

More error types, customization, features etc; PRs & issues are welcome!

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). Show some love to Go errors like our gopher lady here!