<p align="center"><img src="https://user-images.githubusercontent.com/1092882/87815217-d864a680-c882-11ea-9c94-24b67f7125fe.png" alt="errors gopher" width="256px"/></p>

[![Build Status](https://travis-ci.org/bnkamalesh/errors.svg?branch=master)](https://travis-ci.org/bnkamalesh/errors)
[![codecov](https://codecov.io/gh/bnkamalesh/errors/branch/master/graph/badge.svg)](https://codecov.io/gh/bnkamalesh/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/bnkamalesh/errors)](https://goreportcard.com/report/github.com/bnkamalesh/errors)
[![Maintainability](https://api.codeclimate.com/v1/badges/a86629ab167695d4db7f/maintainability)](https://codeclimate.com/github/bnkamalesh/errors)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](https://pkg.go.dev/github.com/bnkamalesh/errors?tab=doc)
[![](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#error-handling)

# Errors

Errors package is a drop-in replacement of the built-in Go errors package with no external dependencies. It lets you create errors of 11 different types which should handle most of the use cases. Some of them are a bit too specific for web applications, but useful nonetheless. Following are the primary features of this package:

1. Multiple (11) error types
2. User friendly message
3. File & line number prefixed to errors
4. HTTP status code and user friendly message (wrapped messages are concatenated) for all error types
5. Helper functions to generate each error type
6. Helper function to get error Type, error type as int, check if error type is wrapped anywhere in chain

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
import(
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
    fmt.Println(err)
    _,msg,_ := errors.HTTPStatusCodeMessage(err)
    fmt.Println(msg)
}
```

### File & line number prefixed to errors

A common annoyance with Go errors which most people are aware of is, figuring out the origin of the error, especially when there are nested function calls. Ever since error annotation was introduced in Go, a lot of people have tried using it to trace out an errors origin by giving function names, contextual message etc in it. e.g. `fmt.Errorf("database query returned error %w", err)`. This errors package, whenever you call the Go error interface's `Error() string` function, it'll print the error prefixed by the filepath and line number. It'd look like `../Users/JohnDoe/apps/main.go:50 hello world` where 'hello world' is the error message.

### HTTP status code & message

The function `errors.HTTPStatusCodeMessage(error) (int, string, bool)` returns the HTTP status code, message, and a boolean value. The boolean is true, if the error is of type *Error from this package. 
If error is nested with multiple errors, it loops through all the levels and returns a single concatenated message. This is illustrated in the 'How to use?' section

## How to use?

Before that, over the years I have tried error with stack trace, annotation, custom error package with error codes etc. Finally, I think this package gives the best of all worlds, for most generic usecases.

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

## Benchmark

Benchmark run on:
<p><img width="320" alt="Screenshot 2020-07-18 at 6 25 22 PM" src="https://user-images.githubusercontent.com/1092882/87852981-241b5c80-c924-11ea-9d22-296acdead7cc.png"></p>

Results
```bash
$ go version
go version go1.14.4 darwin/amd64
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/bnkamalesh/errors
Benchmark_Internal-8                            	 1874256	       639 ns/op	     368 B/op	       5 allocs/op
Benchmark_InternalErr-8                         	 1612707	       755 ns/op	     368 B/op	       5 allocs/op
Benchmark_InternalGetError-8                    	 1700966	       706 ns/op	     464 B/op	       6 allocs/op
Benchmark_InternalGetErrorWithNestedError-8     	 1458368	       823 ns/op	     464 B/op	       6 allocs/op
Benchmark_InternalGetMessage-8                  	 1866562	       643 ns/op	     368 B/op	       5 allocs/op
Benchmark_InternalGetMessageWithNestedError-8   	 1656597	       770 ns/op	     400 B/op	       6 allocs/op
Benchmark_HTTPStatusCodeMessage-8               	26003678	        46.1 ns/op	      16 B/op	       1 allocs/op
BenchmarkHasType-8                              	84689433	        14.2 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/bnkamalesh/errors	14.478s
```

## Contributing

More error types, customization etc; PRs & issues are welcome!

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). Show some love to Go errors like our gopher lady here!