<p align="center"><img src="https://user-images.githubusercontent.com/1092882/87815217-d864a680-c882-11ea-9c94-24b67f7125fe.png" alt="errors gopher" width="256px"/></p>

[![Build Status](https://travis-ci.org/bnkamalesh/errors.svg?branch=master)](https://travis-ci.org/bnkamalesh/errors)
[![codecov](https://codecov.io/gh/bnkamalesh/errors/branch/master/graph/badge.svg)](https://codecov.io/gh/bnkamalesh/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/bnkamalesh/errors)](https://goreportcard.com/report/github.com/bnkamalesh/errors)
[![Maintainability](https://api.codeclimate.com/v1/badges/a86629ab167695d4db7f/maintainability)](https://codeclimate.com/github/bnkamalesh/errors)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](https://pkg.go.dev/github.com/bnkamalesh/errors?tab=doc)
[![](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#error-handling)

# Errors v0.9.4

Errors package is a drop-in replacement of the built-in Go errors package. It lets you create errors of 11 different types,
which should handle most of the use cases. Some of them are a bit too specific for web applications, but useful nonetheless.

Features of this package:

1. Multiple (11) error types
2. Easy handling of User friendly message(s)
3. Stacktrace - formatted, unfromatted, custom format (refer tests in errors_test.go)
4. Retrieve the Program Counters for the stacktrace
5. Retrieve runtime.Frames using `errors.RuntimeFrames(err error)` for the stacktrace
6. HTTP status code and user friendly message (wrapped messages are concatenated) for all error types
7. Helper functions to generate each error type
8. Helper function to get error Type, error type as int, check if error type is wrapped anywhere in chain
9. `fmt.Formatter` support

In case of nested errors, the messages & errors are also looped through the full chain of errors.

### Prerequisites

Go 1.13+

### Available error types

1. TypeInternal - For internal system error. e.g. Database errors
2. TypeValidation - For validation error. e.g. invalid email address
3. TypeInputBody - For invalid input data. e.g. invalid JSON
4. TypeDuplicate - For duplicate content error. e.g. user with email already exists (when trying to register a new user)
5. TypeUnauthenticated - For not authenticated error
6. TypeUnauthorized - For unauthorized access error
7. TypeEmpty - For when an expected non-empty resource, is empty
8. TypeNotFound - For expected resource not found. e.g. user ID not found
9. TypeMaximumAttempts - For attempting the same action more than an allowed threshold
10. TypeSubscriptionExpired - For when a user's 'paid' account has expired
11. TypeDownstreamDependencyTimedout - For when a request to a downstream dependent service times out

Helper functions are available for all the error types. Each of them have 2 helper functions, one which accepts only a string,
and the other which accepts an original error as well as a user friendly message.

All the dedicated error type functions are documented [here](https://pkg.go.dev/github.com/bnkamalesh/errors?tab=doc#DownstreamDependencyTimedout).
Names are consistent with the error type, e.g. errors.Internal(string) and errors.InternalErr(error, string)

### User friendly messages

More often than not when writing APIs, we'd want to respond with an easier to undersand user friendly message.
Instead of returning the raw error and log the raw error.

There are helper functions for all the error types. When in need of setting a friendly message, there
are helper functions with the _suffix_ **'Err'**. All such helper functions accept the original error and a string.

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

err.Error(): /Users/k.balakumaran/go/src/github.com/bnkamalesh/errors/cmd/main.go:16: bar is not happy
hello world!bar is not happy

formatted +v: /Users/k.balakumaran/go/src/github.com/bnkamalesh/errors/cmd/main.go:16: bar is not happy
hello world!bar is not happy

formatted v: bar is not happy

formatted +s: bar is not happy: hello world!

formatted s: bar is not happy

msg: bar is not happy
```

[Playground link](https://go.dev/play/p/-WzDH46f_U5)

### File & line number prefixed to errors

A common annoyance with Go errors which most people are aware of is, figuring out the origin of the error, especially when there are nested function calls. Ever since error annotation was introduced in Go, a lot of people have tried using it to trace out an errors origin by giving function names, contextual message etc in it. e.g. `fmt.Errorf("database query returned error %w", err)`. However this errors package, whenever you call the Go error interface's `Error() string` function, prints the error prefixed by the filepath and line number. It'd look like `../Users/JohnDoe/apps/main.go:50 hello world` where 'hello world' is the error message.

### HTTP status code & message

The function `errors.HTTPStatusCodeMessage(error) (int, string, bool)` returns the HTTP status code, message, and a boolean value. The boolean is true, if the error is of type \*Error from this package. If error is nested, it unwraps and returns a single concatenated message. Sample described in the 'How to use?' section

## How to use?

A sample was already shown in the user friendly message section, following one would show a few more scenarios.

```golang
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/webgo/v6"
	"github.com/bnkamalesh/webgo/v6/middleware/accesslog"
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
		{
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
	}, routes()...)

	router.UseOnSpecialHandlers(accesslog.AccessLog)
	router.Use(accesslog.AccessLog)
	router.Start()
}
```

[webgo](https://github.com/bnkamalesh/webgo) was used to illustrate the usage of the function, `errors.HTTPStatusCodeMessage`. It returns the appropriate http status code, user friendly message stored within, and a boolean value. Boolean value is `true` if the returned error of type \*Error.
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

## Benchmark [2022-01-12]

MacBook Pro (13-inch, 2020, Four Thunderbolt 3 ports), 32 GB 3733 MHz LPDDR4X

```bash
$ go version
go version go1.19.5 darwin/amd64

$ go test -benchmem -bench .
goos: darwin
goarch: amd64
pkg: github.com/bnkamalesh/errors
cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
Benchmark_Internal-8                            	 1526194	       748.8 ns/op	    1104 B/op	       2 allocs/op
Benchmark_Internalf-8                           	 1281465	       944.0 ns/op	    1128 B/op	       3 allocs/op
Benchmark_InternalErr-8                         	 1494351	       806.7 ns/op	    1104 B/op	       2 allocs/op
Benchmark_InternalGetError-8                    	  981162	      1189 ns/op	    1528 B/op	       6 allocs/op
Benchmark_InternalGetErrorWithNestedError-8     	  896322	      1267 ns/op	    1544 B/op	       6 allocs/op
Benchmark_InternalGetMessage-8                  	 1492812	       804.2 ns/op	    1104 B/op	       2 allocs/op
Benchmark_InternalGetMessageWithNestedError-8   	 1362092	       886.3 ns/op	    1128 B/op	       3 allocs/op
Benchmark_HTTPStatusCodeMessage-8               	27494096	        41.38 ns/op	      16 B/op	       1 allocs/op
BenchmarkHasType-8                              	100000000	        10.50 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/bnkamalesh/errors	15.006s
```

## Contributing

More error types, customization, features, multi-errors; PRs & issues are welcome!

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). Show some love to Go errors like our gopher lady here!
