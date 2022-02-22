package webgo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// urlchars is regex to validate characters in a URI parameter
// const urlchars = `([a-zA-Z0-9\*\-+._~!$()=&',;:@%]+)`
// Regex prepared based on http://stackoverflow.com/a/4669750/1359163,
// https://tools.ietf.org/html/rfc3986
// Though the current one allows invalid characters in the URI parameter, it has better performance.
const (
	urlchars                  = `([^/]+)`
	urlwildcard               = `(.*)[^/]`
	urlwildcardWithTrailslash = `(.*)[/]?`
	trailingSlash             = `[/]?`
	errDuplicateKey           = `Error: Duplicate URI keys found`
)

var (
	validHTTPMethods = []string{
		http.MethodOptions,
		http.MethodHead,
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	ctxPool = &sync.Pool{
		New: func() interface{} {
			return new(ContextPayload)
		},
	}
	crwPool = &sync.Pool{
		New: func() interface{} {
			return new(customResponseWriter)
		},
	}
)

// customResponseWriter is a custom HTTP response writer
type customResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	written       bool
	headerWritten bool
}

// WriteHeader is the interface implementation to get HTTP response code and add
// it to the custom response writer
func (crw *customResponseWriter) WriteHeader(code int) {
	if crw.headerWritten {
		return
	}

	crw.headerWritten = true
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// Write is the interface implementation to respond to the HTTP request,
// but check if a response was already sent.
func (crw *customResponseWriter) Write(body []byte) (int, error) {
	crw.WriteHeader(crw.statusCode)
	crw.written = true
	return crw.ResponseWriter.Write(body)
}

// Flush calls the http.Flusher to clear/flush the buffer
func (crw *customResponseWriter) Flush() {
	if rw, ok := crw.ResponseWriter.(http.Flusher); ok {
		rw.Flush()
	}
}

// Hijack implements the http.Hijacker interface
func (crw *customResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := crw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}

	return nil, nil, errors.New("unable to create hijacker")
}

// CloseNotify implements the http.CloseNotifier interface
func (crw *customResponseWriter) CloseNotify() <-chan bool {
	if n, ok := crw.ResponseWriter.(http.CloseNotifier); ok {
		return n.CloseNotify()
	}
	return nil
}

func (crw *customResponseWriter) reset() {
	crw.statusCode = 0
	crw.written = false
	crw.headerWritten = false
	crw.ResponseWriter = nil
}

// Middleware is the signature of WebGo's middleware
type Middleware func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// discoverRoute returns the correct 'route', for the given request
func discoverRoute(path string, routes []*Route) (*Route, map[string]string) {
	for _, route := range routes {
		if ok, params := route.matchPath(path); ok {
			return route, params
		}
	}
	return nil, nil
}

// Router is the HTTP router
type Router struct {
	optHandlers    []*Route
	headHandlers   []*Route
	getHandlers    []*Route
	postHandlers   []*Route
	putHandlers    []*Route
	patchHandlers  []*Route
	deleteHandlers []*Route
	allHandlers    map[string][]*Route

	// NotFound is the generic handler for 404 resource not found response
	NotFound http.HandlerFunc

	// NotImplemented is the generic handler for 501 method not implemented
	NotImplemented http.HandlerFunc

	// config has all the app config
	config *Config

	// httpServer is the server handler for the active HTTP server
	httpServer *http.Server
	// httpsServer is the server handler for the active HTTPS server
	httpsServer *http.Server
}

// methodRoutes returns the list of Routes handling the HTTP method given the request
func (rtr *Router) methodRoutes(method string) (routes []*Route) {
	switch method {
	case http.MethodOptions:
		return rtr.optHandlers
	case http.MethodHead:
		return rtr.headHandlers
	case http.MethodGet:
		return rtr.getHandlers
	case http.MethodPost:
		return rtr.postHandlers
	case http.MethodPut:
		return rtr.putHandlers
	case http.MethodPatch:
		return rtr.patchHandlers
	case http.MethodDelete:
		return rtr.deleteHandlers
	}

	return nil
}

func (rtr *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// a custom response writer is used to set appropriate HTTP status code in case of
	// encoding errors. i.e. if there's a JSON encoding issue while responding,
	// the HTTP status code would say 200, and and the JSON payload {"status": 500}
	crw := newCRW(rw, http.StatusOK)

	routes := rtr.methodRoutes(r.Method)
	if routes == nil {
		// serve 501 when HTTP method is not implemented
		crw.statusCode = http.StatusNotImplemented
		rtr.NotImplemented(crw, r)
		releaseCRW(crw)
		return
	}

	path := r.URL.EscapedPath()
	route, params := discoverRoute(path, routes)
	if route == nil {
		// serve 404 when there are no matching routes
		crw.statusCode = http.StatusNotFound
		rtr.NotFound(crw, r)
		releaseCRW(crw)
		return
	}

	ctxPayload := newContext()
	ctxPayload.Route = route
	ctxPayload.URIParams = params

	// webgo context is injected to the HTTP request context
	*r = *r.WithContext(
		context.WithValue(
			r.Context(),
			wgoCtxKey,
			ctxPayload,
		),
	)

	route.serve(crw, r)

	// IMPORTANT/TODO: if the handler panics, resources are not released from the pool.
	// `defer releasePoolResources(crw, ctxPayload)` would solve the problem but with a potential performance hit
	releasePoolResources(crw, ctxPayload)
}

// Use adds a middleware layer
func (rtr *Router) Use(mm ...Middleware) {
	for _, handlers := range rtr.allHandlers {
		for idx := range handlers {
			route := handlers[idx]
			if route.skipMiddleware {
				continue
			}

			route.use(mm...)
		}
	}
}

// UseOnSpecialHandlers adds middleware to the 2 special handlers of webgo
func (rtr *Router) UseOnSpecialHandlers(mm ...Middleware) {
	// v3.2.1 introduced the feature of adding middleware to both notfound & not implemented
	// handlers
	/*
		- It was added considering an `accesslog` middleware, where all requests should be logged
		# This is now being moved to a separate function considering an authentication middleware, where all requests
		  including 404 & 501 would respond with `not authenticated` if you do not have special handling
		  within the middleware. It is a cleaner implementation to avoid this and let users add their
		  middleware separately to NOTFOUND & NOTIMPLEMENTED handlers
	*/

	for idx := range mm {
		m := mm[idx]
		nf := rtr.NotFound
		rtr.NotFound = func(rw http.ResponseWriter, req *http.Request) {
			m(rw, req, nf)
		}

		ni := rtr.NotImplemented
		rtr.NotImplemented = func(rw http.ResponseWriter, req *http.Request) {
			m(rw, req, ni)
		}
	}
}

// Add is a convenience method used to add a new route to an already initialized router
// Important: `.Use` should be used only after all routes are added
func (rtr *Router) Add(routes ...*Route) {
	hmap := httpHandlers(routes)
	rtr.optHandlers = append(rtr.optHandlers, hmap[http.MethodOptions]...)
	rtr.headHandlers = append(rtr.headHandlers, hmap[http.MethodHead]...)
	rtr.getHandlers = append(rtr.getHandlers, hmap[http.MethodGet]...)
	rtr.postHandlers = append(rtr.postHandlers, hmap[http.MethodPost]...)
	rtr.putHandlers = append(rtr.putHandlers, hmap[http.MethodPut]...)
	rtr.patchHandlers = append(rtr.patchHandlers, hmap[http.MethodPatch]...)
	rtr.deleteHandlers = append(rtr.deleteHandlers, hmap[http.MethodDelete]...)

	all := rtr.allHandlers
	if all == nil {
		all = map[string][]*Route{}
	}

	for _, key := range supportedHTTPMethods {
		newlist, hasKey := hmap[key]
		if !hasKey {
			continue
		}
		if all[key] == nil {
			all[key] = make([]*Route, 0, len(hmap))
		}
		all[key] = append(all[key], newlist...)
	}

	rtr.allHandlers = all
}

func newCRW(rw http.ResponseWriter, rCode int) *customResponseWriter {
	crw := crwPool.Get().(*customResponseWriter)
	crw.ResponseWriter = rw
	crw.statusCode = rCode
	return crw
}

func releaseCRW(crw *customResponseWriter) {
	crw.reset()
	crwPool.Put(crw)
}

func newContext() *ContextPayload {
	return ctxPool.Get().(*ContextPayload)
}

func releaseContext(cp *ContextPayload) {
	cp.reset()
	ctxPool.Put(cp)
}

func releasePoolResources(crw *customResponseWriter, cp *ContextPayload) {
	releaseCRW(crw)
	releaseContext(cp)
}

// NewRouter initializes & returns a new router instance with all the configurations and routes set
func NewRouter(cfg *Config, routes ...*Route) *Router {
	r := &Router{
		NotFound: http.NotFound,
		NotImplemented: func(rw http.ResponseWriter, req *http.Request) {
			Send(rw, "", "501 Not Implemented", http.StatusNotImplemented)
		},
		config: cfg,
	}

	r.Add(routes...)

	return r
}

// checkDuplicateRoutes checks if any of the routes have duplicate name or URI pattern
func checkDuplicateRoutes(idx int, route *Route, routes []*Route) {
	// checking if the URI pattern is duplicated
	for i := 0; i < idx; i++ {
		rt := routes[i]

		if rt.Name == route.Name {
			LOGHANDLER.Info(
				fmt.Sprintf(
					"Duplicate route name('%s') detected",
					rt.Name,
				),
			)
		}

		if rt.Method != route.Method {
			continue
		}

		// regex pattern match
		if ok, _ := rt.matchPath(route.Pattern); !ok {
			continue
		}

		LOGHANDLER.Warn(
			fmt.Sprintf(
				"Duplicate URI pattern detected.\nPattern: '%s'\nDuplicate pattern: '%s'",
				rt.Pattern,
				route.Pattern,
			),
		)
		LOGHANDLER.Warn("Only the first route to match the URI pattern would handle the request")
	}
}

// httpHandlers returns all the handlers in a map, for each HTTP method
func httpHandlers(routes []*Route) map[string][]*Route {
	handlers := map[string][]*Route{}

	handlers[http.MethodHead] = []*Route{}
	handlers[http.MethodGet] = []*Route{}

	for idx, route := range routes {
		found := false
		for _, validMethod := range validHTTPMethods {
			if route.Method == validMethod {
				found = true
				break
			}
		}

		if !found {
			LOGHANDLER.Fatal(
				fmt.Sprintf(
					"Unsupported HTTP method provided. Method: '%s'",
					route.Method,
				),
			)
			return nil
		}

		if route.Handlers == nil || len(route.Handlers) == 0 {
			LOGHANDLER.Fatal(
				fmt.Sprintf(
					"No handlers provided for the route '%s', method '%s'",
					route.Pattern,
					route.Method,
				),
			)
			return nil
		}

		err := route.init()
		if err != nil {
			LOGHANDLER.Fatal("Unsupported URI pattern.", route.Pattern, err)
			return nil
		}

		checkDuplicateRoutes(idx, route, routes)

		handlers[route.Method] = append(handlers[route.Method], route)
	}

	return handlers
}
