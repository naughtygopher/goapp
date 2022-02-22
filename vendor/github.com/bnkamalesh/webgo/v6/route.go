package webgo

import (
	"fmt"
	"net/http"
	"strings"
)

// Route defines a route for each API
type Route struct {
	// Name is unique identifier for the route
	Name string
	// Method is the HTTP request method/type
	Method string
	// Pattern is the URI pattern to match
	Pattern string
	// TrailingSlash if set to true, the URI will be matched with or without
	// a trailing slash. IMPORTANT: It does not redirect.
	TrailingSlash bool

	// FallThroughPostResponse if enabled will execute all the handlers even if a response was already sent to the client
	FallThroughPostResponse bool

	// Handlers is a slice of http.HandlerFunc which can be middlewares or anything else. Though only 1 of them will be allowed to respond to client.
	// subsequent writes from the following handlers will be ignored
	Handlers []http.HandlerFunc

	hasWildcard bool
	parts       []routePart

	// skipMiddleware if true, middleware added using `router` will not be applied to this Route.
	// This is used only when a Route is set using the RouteGroup, which can have its own set of middleware
	skipMiddleware bool

	initialized bool

	serve http.HandlerFunc
}
type routePart struct {
	isVariable  bool
	hasWildcard bool
	// part will be the key name, if it's a variable/named URI parameter
	part string
}

func (r *Route) parseURIWithParams() {
	// if there are no URI params, then there's no need to set route parts
	if !strings.Contains(r.Pattern, ":") {
		return
	}

	parts := strings.Split(r.Pattern, "/")
	if len(parts) == 1 {
		return
	}

	rparts := make([]routePart, 0, len(parts))
	for _, part := range parts[1:] {
		hasParam := false
		hasWildcard := false

		if strings.Contains(part, ":") {
			hasParam = true
		}
		if strings.Contains(part, "*") {
			r.hasWildcard = true
			hasWildcard = true
		}

		key := strings.ReplaceAll(part, ":", "")
		key = strings.ReplaceAll(key, "*", "")
		rparts = append(
			rparts,
			routePart{
				isVariable:  hasParam,
				hasWildcard: hasWildcard,
				part:        key,
			})
	}
	r.parts = rparts
}

// init prepares the URIKeys, compile regex for the provided pattern
func (r *Route) init() error {
	if r.initialized {
		return nil
	}
	r.parseURIWithParams()
	r.serve = defaultRouteServe(r)

	r.initialized = true
	return nil
}

// matchPath matches the requestURI with the URI pattern of the route.
// If the path is an exact match (i.e. no URI parameters), then the second parameter ('isExactMatch') is true
func (r *Route) matchPath(requestURI string) (bool, map[string]string) {
	p := r.Pattern
	if r.TrailingSlash {
		p += "/"
	} else {
		if requestURI[len(requestURI)-1] == '/' {
			return false, nil
		}
	}
	if r.Pattern == requestURI || p == requestURI {
		return true, nil
	}

	return r.matchWithWildcard(requestURI)
}

func (r *Route) matchWithWildcard(requestURI string) (bool, map[string]string) {
	params := map[string]string{}
	uriParts := strings.Split(requestURI, "/")[1:]

	partsLastIdx := len(r.parts) - 1
	partIdx := 0
	paramParts := make([]string, 0, len(uriParts))
	for idx, part := range uriParts {
		// if part is empty, it means it's end of URI with trailing slash
		if part == "" {
			break
		}

		if partIdx > partsLastIdx {
			return false, nil
		}

		currentPart := r.parts[partIdx]
		if !currentPart.isVariable && currentPart.part != part {
			return false, nil
		}

		paramParts = append(paramParts, part)
		if currentPart.isVariable {
			params[currentPart.part] = strings.Join(paramParts, "/")
		}

		if !currentPart.hasWildcard {
			paramParts = make([]string, 0, len(uriParts)-idx)
			partIdx++
			continue
		}

		nextIdx := partIdx + 1
		if nextIdx > partsLastIdx {
			continue
		}
		nextPart := r.parts[nextIdx]

		// if the URI has more parts/params after wildcard,
		// the immediately following part after wildcard cannot be a variable or another wildcard.
		if !nextPart.isVariable && nextPart.part == part {
			// remove the last added 'part' from parameters, as it's part of the static URI
			params[currentPart.part] = strings.Join(paramParts[:len(paramParts)-1], "/")
			paramParts = make([]string, 0, len(uriParts)-idx)
			partIdx += 2
		}
	}

	return true, params
}

func (r *Route) use(mm ...Middleware) {
	for idx := range mm {
		m := mm[idx]
		srv := r.serve
		r.serve = func(rw http.ResponseWriter, req *http.Request) {
			m(rw, req, srv)
		}
	}
}

func routeServeChainedHandlers(r *Route) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		crw, ok := rw.(*customResponseWriter)
		if !ok {
			crw = newCRW(rw, http.StatusOK)
		}

		for _, handler := range r.Handlers {
			if crw.written && !r.FallThroughPostResponse {
				break
			}
			handler(crw, req)
		}
	}
}

func defaultRouteServe(r *Route) http.HandlerFunc {
	if len(r.Handlers) > 1 {
		return routeServeChainedHandlers(r)
	}

	// when there is only 1 handler, custom response writer is not required to check if response
	// is already written or fallthrough is enabled
	return r.Handlers[0]
}

type RouteGroup struct {
	routes []*Route
	// skipRouterMiddleware if set to true, middleware applied to the router will not be applied
	// to this route group.
	skipRouterMiddleware bool
	// PathPrefix is the URI prefix for all routes in this group
	PathPrefix string
}

func (rg *RouteGroup) Add(rr ...Route) {
	for idx := range rr {
		route := rr[idx]
		route.skipMiddleware = rg.skipRouterMiddleware
		route.Pattern = fmt.Sprintf("%s%s", rg.PathPrefix, route.Pattern)
		_ = route.init()
		rg.routes = append(rg.routes, &route)
	}
}

func (rg *RouteGroup) Use(mm ...Middleware) {
	for idx := range rg.routes {
		route := rg.routes[idx]
		route.use(mm...)
	}
}

func (rg *RouteGroup) Routes() []*Route {
	return rg.routes
}

func NewRouteGroup(pathPrefix string, skipRouterMiddleware bool, rr ...Route) *RouteGroup {
	rg := RouteGroup{
		PathPrefix:           pathPrefix,
		skipRouterMiddleware: skipRouterMiddleware,
	}
	rg.Add(rr...)
	return &rg
}
