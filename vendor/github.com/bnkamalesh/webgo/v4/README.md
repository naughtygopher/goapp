<p align="center"><img src="https://user-images.githubusercontent.com/1092882/60883564-20142380-a268-11e9-988a-d98fb639adc6.png" alt="webgo gopher" width="256px"/></p>

[![](https://travis-ci.org/bnkamalesh/webgo.svg?branch=master)](https://travis-ci.org/bnkamalesh/webgo)
[![coverage](https://img.shields.io/codecov/c/github/bnkamalesh/webgo.svg)](https://codecov.io/gh/bnkamalesh/webgo)
[![](https://goreportcard.com/badge/github.com/bnkamalesh/webgo)](https://goreportcard.com/report/github.com/bnkamalesh/webgo)
[![](https://api.codeclimate.com/v1/badges/85b3a55c3fa6b4c5338d/maintainability)](https://codeclimate.com/github/bnkamalesh/webgo/maintainability)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](http://godoc.org/github.com/bnkamalesh/webgo)
[![](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#web-frameworks)

# WebGo v4.1.1

WebGo is a minimalistic framework for [Go](https://golang.org) to build web applications (server side) with zero 3rd party dependencies. Unlike full-fledged frameworks, it gets out of your way as soon as possible in the execution flow. WebGo has always been and will always be Go standard library compliant; with the HTTP handlers having the same signature as [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc).

### Important ❗

- Regression introduced in `v4.0.4`, exists on `v4.0.6` as well. Requests panic when not using Webgo's response methods (e.g. R200, Send, SendResponse etc.) because the default HTTP status is set as 0

- `ContextPayload.URIParams(*http.Request)map[string]string` was replaced despite being newly introduced in v3.5.4. The new function is `ContextPayload.Params()map[string]string`, and has a slight performance advantage compared to URIParams


### Index

1. [Router](https://github.com/bnkamalesh/webgo#router)
2. [Handler chaining](https://github.com/bnkamalesh/webgo#handler-chaining)
3. [Middleware](https://github.com/bnkamalesh/webgo#middleware)
4. [Helper functions](https://github.com/bnkamalesh/webgo#helper-functions)
5. [HTTPS ready](https://github.com/bnkamalesh/webgo#https-ready)
6. [Graceful shutdown](https://github.com/bnkamalesh/webgo#graceful-shutdown)
7. [Logging](https://github.com/bnkamalesh/webgo#logging)
8. [Usage](https://github.com/bnkamalesh/webgo#usage)


## Router


The router is one of the most important component of a web application. It helps identify the HTTP requests and pass them on to respective handlers. A handler is identified using a [URI](https://developer.mozilla.org/en-US/docs/Glossary/URI). WebGo supports defining URIs with the following patterns

1. `/api/users` 
	- Static URI pattern with no variables
2. `/api/users/:userID` 
	- URI pattern with variable `userID` (named URI parameter)
	- This will **_not_** match `/api/users/johndoe/account`. It only matches till `/api/users/johndoe/`
		- If TrailingSlash is set to true, refer to [sample](https://github.com/bnkamalesh/webgo#sample)
3. `/api/users/:misc*`
	- Named URI variable `misc`
	- This matches everything after `/api/users`. e.g. `/api/users/a/b/c/d`

If multiple patterns match the same URI, the first matching handler would be executed. Refer to the [sample](https://github.com/bnkamalesh/webgo#sample) to see how routes are configured. A WebGo [Route](https://godoc.org/github.com/bnkamalesh/webgo#Route) is defined as following:

```golang
webgo.Route{
	// A name for the API (preferrably unique)
	Name string
	// HTTP verb, i.e. GET, POST, PUT, PATCH, HEAD, DELETE
	Method string
	// The URI pattern
	Pattern string
	// If the URI ends with a '/', should it be considered valid or not? e.g. '/api/users' vs '/api/users/'
	TrailingSlash bool
	// In case of chained handlers, should the execution continue after one of the handlers have 
	// responded to the HTTP request
	FallThroughPostResponse bool
	// The list of HTTP handlers
	Handlers []http.HandlerFunc
}
```

You can access named parameters of the URI using the `Context` function.

```golang
func helloWorld(w http.ResponseWriter, r *http.Request) {
	// WebGo context
	wctx := webgo.Context(r)
	// URI paramaters, map[string]string
	params := wctx.Params()
	// route, the webgo.Route which is executing this request
	route := wctx.Route
	webgo.R200(
		w,
		fmt.Sprintf(
			"Route name: '%s', params: '%s'", 
			route.Name,
			params, 
			),
	)
}
```

## Handler chaining

Handler chaining lets you execute multiple handlers for a given route. Execution of a chain can be configured to run even after a handler has written a response to the http request. This is made possible by setting `FallThroughPostResponse` to `true` (refer [sample](https://github.com/bnkamalesh/webgo#sample)).

```golang
webgo.Route{
	Name: "chained",
	Method: http.MethodGet,
	Pattern: "/api",
	TrailingSlash: false,
	FallThroughPostResponse: true,
	Handlers []http.HandlerFunc{
		handler1,
		handler2,
		.
		.
		.
	}
}
```

## Middleware

WebGo middleware lets you wrap all the routes with a middleware. Unlike handler chaining, middleware applies to all the handlers. All middleware should be of type [Middlware](https://godoc.org/github.com/bnkamalesh/webgo#Middleware). The router exposes a method [Use](https://godoc.org/github.com/bnkamalesh/webgo#Router.Use) && [UseOnSpecialHandlers](https://godoc.org/github.com/bnkamalesh/webgo#Router.UseOnSpecialHandlers) to add a Middleware to the router. Following code shows how a middleware can be used in WebGo.

```golang
import (
	"github.com/bnkamalesh/webgo/v4"
	"github.com/bnkamalesh/webgo/v4/middleware"
)

func routes() []*webgo.Route {
	return []*webgo.Route{
		&webo.Route{
			Name: "home",
			Method: http.http.MethodGet,
			Pattern: "/",
			Handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					webgo.R200(w, "home")
				}
			},
		},
	}
}

func main() {
	router := webgo.NewRouter(*webgo.Config{
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

Any number of middleware can be added to the router, the order of execution of middleware would be [LIFO](https://en.wikipedia.org/wiki/Stack_(abstract_data_type)) (Last In First Out). i.e. in case of the following code

```golang
func main() {
	router.Use(middleware.AccessLog)
	router.Use(middleware.CorsWrap())
}
```

**_CorsWrap_** would be executed first, followed by **_AccessLog_**.

## Helper functions

WebGo provides a few helper functions.

1. [ResponseStatus(w http.ResponseWriter)](https://godoc.org/github.com/bnkamalesh/webgo#ResponseStatus) get the HTTP status code from response writer
2. [SendHeader(w http.ResponseWriter, rCode int)](https://godoc.org/github.com/bnkamalesh/webgo#SendHeader) - Send only an HTTP response header with the provided response code.
3. [Send(w http.ResponseWriter, contentType string, data interface{}, rCode int)](https://godoc.org/github.com/bnkamalesh/webgo#Send) - Send any response as is, with the provided content type and response code
4. [SendResponse(w http.ResponseWriter, data interface{}, rCode int)](https://godoc.org/github.com/bnkamalesh/webgo#SendResponse) - Send a JSON response wrapped in WebGo's default response struct.
5. [SendError(w http.ResponseWriter, data interface{}, rCode int)](https://godoc.org/github.com/bnkamalesh/webgo#SendError) - Send a JSON response wrapped in WebGo's default error response struct
6. [Render(w http.ResponseWriter, data interface{}, rCode int, tpl *template.Template)](https://godoc.org/github.com/bnkamalesh/webgo#Render) - Render renders a Go template, with the provided data & response code.

Few more helper functions are available, you can check them [here](https://godoc.org/github.com/bnkamalesh/webgo#R200). 

When using `Send` or `SendResponse`, the response is wrapped in WebGo's [response struct](https://github.com/bnkamalesh/webgo/blob/master/responses.go#L17) and is serialized as JSON.

```json
{
	"data": "<any valid JSON payload>",
	"status": "<HTTP status code, of type integer>"
}
```

When using `SendError`, the response is wrapped in WebGo's [error response struct](https://github.com/bnkamalesh/webgo/blob/master/responses.go#L23) and is serialzied as JSON.

```json
{
	"errors": "<any valid JSON payload>",
	"status": "<HTTP status code, of type integer>"
}
```

## HTTPS ready

HTTPS server can be started easily, by providing the key & cert file. You can also have both HTTP & HTTPS servers running side by side. 

Start HTTPS server

```golang
cfg := &webgo.Config{
	Port: "80",
	HTTPSPort: "443",
	CertFile: "/path/to/certfile",
	KeyFile: "/path/to/keyfile",
}
router := webgo.NewRouter(cfg, routes())
router.StartHTTPS()
```

Starting both HTTP & HTTPS server

```golang
cfg := &webgo.Config{
	Port: "80",
	HTTPSPort: "443",
	CertFile: "/path/to/certfile",
	KeyFile: "/path/to/keyfile",
}

router := webgo.NewRouter(cfg, routes())
go router.StartHTTPS()
router.Start()
```

## Graceful shutdown

Graceful shutdown lets you shutdown the server without affecting any live connections/clients connected to the server. It will complete executing all the active/live requests before shutting down.

Sample code to show how to use shutdown

```golang
func main() {
	osSig := make(chan os.Signal, 5)

	cfg := &webgo.Config{
		Host:            "",
		Port:            "8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    60 * time.Second,
		ShutdownTimeout: 15 * time.Second,
	}
	router := webgo.NewRouter(cfg, routes())

	go func() {
		<-osSig
		// Initiate HTTP server shutdown
		err := router.Shutdown()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println("shutdown complete")
			os.Exit(0)
		}

		// If you have HTTPS server running, you can use the following code
		// err := router.ShutdownHTTPS()
		// if err != nil {
		// 	fmt.Println(err)
		// 	os.Exit(1)
		// } else {
		// 	fmt.Println("shutdown complete")
		// 	os.Exit(0)
		// }
	}()

	signal.Notify(osSig, os.Interrupt, syscall.SIGTERM)

	router.Start()

	for {
		// Prevent main thread from exiting, and wait for shutdown to complete
		time.Sleep(time.Second * 1)
	}
}
```

## Logging

WebGo exposes a singleton & global scoped logger variable [LOGHANDLER](https://godoc.org/github.com/bnkamalesh/webgo#Logger) with which you can plugin your custom logger. Any custom logger should implement WebGo's [Logger](https://godoc.org/github.com/bnkamalesh/webgo#Logger) interface.

```golang
type Logger interface {
    Debug(data ...interface{})
    Info(data ...interface{})
    Warn(data ...interface{})
    Error(data ...interface{})
    Fatal(data ...interface{})
}
```

### Configuring the default Logger

The default logger uses Go standard library's `log.Logger` with `os.Stdout` (for debug and info logs) & `os.Stderr` (for warning, error, fatal) as default io.Writers. You can set the io.Writer as well as disable specific types of logs using the `GlobalLoggerConfig(stdout, stderr, cfgs...)` function.

```golang
GlobalLoggerConfig(nil, nil, LogCfgDisableDebug, LogCfgDisableInfo...)
```

Usage is shown in `cmd/main.go`.

## Usage

A fully functional sample is provided [here](https://github.com/bnkamalesh/webgo/blob/master/cmd/main.go). You can try the following API calls with the sample app.

1. `http://localhost:8080/`
	- Route with no named parameters configured
2. `http://localhost:8080/matchall/`
	- Route with wildcard parameter configured
	- All URIs which begin with `/matchall` will be matched because it has a wildcard variable
	- e.g. 
		- http://localhost:8080/matchall/hello
		- http://localhost:8080/matchall/hello/world
		- http://localhost:8080/matchall/hello/world/user
3. `http://localhost:8080/api/<param>
	- Route with a named 'param' configured
	- It will match all requests which match `/api/<single parameter>`
	- e.g.
		- http://localhost:8080/api/hello
		- http://localhost:8080/api/world

### How to run the sample

If you have Go installed on your system, open your terminal and:

```bash
$ cd $GOPATH/src
$ mkdir -p github.com/bnkamalesh
$ cd github.com/bnkamalesh
$ git clone https://github.com/bnkamalesh/webgo.git
$ cd webgo
$ go run cmd/main.go

Info 2020/06/03 12:55:26 HTTP server, listening on :8080
```

Or if you have [Docker](https://www.docker.com/), open your terminal and:

```bash
$ git clone https://github.com/bnkamalesh/webgo.git
$ cd webgo
$ docker run \
-p 8080:8080 \
-v ${PWD}:/go/src/github.com/bnkamalesh/webgo/ \
-w /go/src/github.com/bnkamalesh/webgo/cmd \
--rm -ti golang:latest go run main.go

Info 2020/06/03 12:55:26 HTTP server, listening on :8080
```

### Contributing

Refer [here](https://github.com/bnkamalesh/webgo/blob/master/CONTRIBUTING.md) to find out details about making a contribution

### Credits

Thanks to all the [contributors](https://github.com/bnkamalesh/webgo/graphs/contributors)

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). WebGo stays out of developers' way, so sitback and enjoy a cup of coffee like this gopher.
