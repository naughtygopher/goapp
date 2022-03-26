<p align="center"><img src="https://user-images.githubusercontent.com/1092882/60883564-20142380-a268-11e9-988a-d98fb639adc6.png" alt="webgo gopher" width="256px"/></p>

[![](https://travis-ci.org/bnkamalesh/webgo.svg?branch=master)](https://travis-ci.org/bnkamalesh/webgo)
[![coverage](https://img.shields.io/codecov/c/github/bnkamalesh/webgo.svg)](https://codecov.io/gh/bnkamalesh/webgo)
[![](https://goreportcard.com/badge/github.com/bnkamalesh/webgo)](https://goreportcard.com/report/github.com/bnkamalesh/webgo)
[![](https://api.codeclimate.com/v1/badges/85b3a55c3fa6b4c5338d/maintainability)](https://codeclimate.com/github/bnkamalesh/webgo/maintainability)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](http://godoc.org/github.com/bnkamalesh/webgo)
[![](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#web-frameworks)

# WebGo v6.6.2

WebGo is a minimalistic framework for [Go](https://golang.org) to build web applications (server side) with no 3rd party dependencies. WebGo will always be Go standard library compliant; with the HTTP handlers having the same signature as [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc).

### Contents

1. [Router](https://github.com/bnkamalesh/webgo#router)
2. [Handler chaining](https://github.com/bnkamalesh/webgo#handler-chaining)
3. [Middleware](https://github.com/bnkamalesh/webgo#middleware)
4. [Error handling](https://github.com/bnkamalesh/webgo#error-handling)
5. [Helper functions](https://github.com/bnkamalesh/webgo#helper-functions)
6. [HTTPS ready](https://github.com/bnkamalesh/webgo#https-ready)
7. [Graceful shutdown](https://github.com/bnkamalesh/webgo#graceful-shutdown)
8. [Logging](https://github.com/bnkamalesh/webgo#logging)
9. [Server-Sent Events](https://github.com/bnkamalesh/webgo#server-sent-events)
10. [Usage](https://github.com/bnkamalesh/webgo#usage)


## Router

Webgo has a simplistic, regex based router and supports defining [URI](https://developer.mozilla.org/en-US/docs/Glossary/URI)s with the following patterns

1. `/api/users` - URI with no dynamic values
2. `/api/users/:userID` 
	- URI with a named parameter, `userID`
	- If TrailingSlash is set to true, it will accept the URI ending with a '/', refer to [sample](https://github.com/bnkamalesh/webgo#sample)
3. `/api/users/:misc*`
	- Named URI parameter `misc`, with a wildcard suffix '*'
	- This matches everything after `/api/users`. e.g. `/api/users/a/b/c/d`

When there are multiple handlers matching the same URI, only the first occurring handler will handle the request.
Refer to the [sample](https://github.com/bnkamalesh/webgo#sample) to see how routes are configured.  You can access named parameters of the URI using the `Context` function. 

Note: webgo Context is **not** available inside the special handlers (not found & method not implemented)

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

Handler chaining lets you execute multiple handlers for a given route. Execution of a chain can be configured to run even after a handler has written a response to the HTTP request, if you set `FallThroughPostResponse` to `true` (refer [sample](https://github.com/bnkamalesh/webgo/blob/master/cmd/main.go#L70)).

## Middleware

WebGo [middlware](https://godoc.org/github.com/bnkamalesh/webgo#Middleware) lets you wrap all the routes with a middleware unlike handler chaining. The router exposes a method [Use](https://godoc.org/github.com/bnkamalesh/webgo#Router.Use) && [UseOnSpecialHandlers](https://godoc.org/github.com/bnkamalesh/webgo#Router.UseOnSpecialHandlers) to add a Middleware to the router.

NotFound && NotImplemented are considered `Special` handlers. `webgo.Context(r)` within special handlers will return `nil`.

Any number of middleware can be added to the router, the order of execution of middleware would be [LIFO](https://en.wikipedia.org/wiki/Stack_(abstract_data_type)) (Last In First Out). i.e. in case of the following code

```golang
func main() {
	router.Use(accesslog.AccessLog, cors.CORS(nil))
	router.Use(<more middleware>)
}
```

**_CorsWrap_** would be executed first, followed by **_AccessLog_**.

## Error handling

Webgo context has 2 methods to [set](https://github.com/bnkamalesh/webgo/blob/master/webgo.go#L60) & [get](https://github.com/bnkamalesh/webgo/blob/master/webgo.go#L66) erro within a request context. It enables Webgo to implement a single middleware where you can handle error returned within an HTTP handler. [set error](https://github.com/bnkamalesh/webgo/blob/master/cmd/main.go#L45), [get error](https://github.com/bnkamalesh/webgo/blob/master/cmd/main.go#L51).

## Helper functions

WebGo provides a few helper functions. When using `Send` or `SendResponse` (other Rxxx responder functions), the response is wrapped in WebGo's [response struct](https://github.com/bnkamalesh/webgo/blob/master/responses.go#L17) and is serialized as JSON.

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
router := webgo.NewRouter(cfg, routes()...)
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

router := webgo.NewRouter(cfg, routes()...)
go router.StartHTTPS()
router.Start()
```

## Graceful shutdown

Graceful shutdown lets you shutdown the server without affecting any live connections/clients connected to the server. Any new connection request after initiating a shutdown would be ignored.

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
	router := webgo.NewRouter(cfg, routes()...)

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

WebGo exposes a singleton & global scoped logger variable [LOGHANDLER](https://godoc.org/github.com/bnkamalesh/webgo#Logger) with which you can plug in your custom logger by implementing the [Logger](https://godoc.org/github.com/bnkamalesh/webgo#Logger) interface.

### Configuring the default Logger

The default logger uses Go standard library's `log.Logger` with `os.Stdout` (for debug and info logs) & `os.Stderr` (for warning, error, fatal) as default io.Writers. You can set the io.Writer as well as disable specific types of logs using the `GlobalLoggerConfig(stdout, stderr, cfgs...)` function.

## Server-Sent Events

[MDN has a very good documentation of what SSE (Server-Sent Events)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) are. The sample app provided shows how to use the SSE extension of webgo.
## Usage

A fully functional sample is provided [here](https://github.com/bnkamalesh/webgo/blob/master/cmd/main.go).
### Benchmark

1. [the-benchmarker](https://github.com/the-benchmarker/web-frameworks)
2. [go-web-framework-benchmark](https://github.com/smallnest/go-web-framework-benchmark)

### Contributing

Refer [here](https://github.com/bnkamalesh/webgo/blob/master/CONTRIBUTING.md) to find out details about making a contribution

### Credits

Thanks to all the [contributors](https://github.com/bnkamalesh/webgo/graphs/contributors)

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). WebGo stays out of developers' way, so sitback and enjoy a cup of coffee.
