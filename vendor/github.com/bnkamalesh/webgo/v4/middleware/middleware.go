// Package middleware defines the signature/type which can be added as a middleware to Webgo.
// It also has a 2 default middleware access logs & CORS handling.
// This package also provides 2 chainable to handlers to handle CORS in individual routes
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bnkamalesh/webgo/v4"
)

// AccessLog is a middleware which prints access log to stdout
func AccessLog(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, req)
	end := time.Now()

	webgo.LOGHANDLER.Info(
		fmt.Sprintf(
			"%s %s %s %s %d",
			end.Format("2006-01-02 15:04:05 -0700 MST"),
			req.Method,
			req.URL.String(),
			end.Sub(start).String(),
			webgo.ResponseStatus(rw),
		),
	)
}

const (
	headerOrigin       = "Access-Control-Allow-Origin"
	headerMethods      = "Access-Control-Allow-Methods"
	headerCreds        = "Access-Control-Allow-Credentials"
	headerAllowHeaders = "Access-Control-Allow-Headers"
	headerReqHeaders   = "Access-Control-Request-Headers"
	headerGetOrigin    = "Origin"
	allowMethods       = "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS"
	allowHeaders       = "Accept,Content-Type,Content-Length,Accept-Encoding,Access-Control-Request-Headers,"
)

func deprecationLog() {
	webgo.LOGHANDLER.Warn("this middleware is deprecated, use github.com/bnkamalesh/middleware/cors")
}

// Cors is a basic CORS middleware which can be added to individual handlers
func Cors(allowedOrigins ...string) http.HandlerFunc {
	deprecationLog()
	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		allowed := false
		// Set appropriate response headers required for CORS
		reqOrigin := req.Header.Get(headerGetOrigin)
		for _, o := range allowedOrigins {
			// Set appropriate response headers required for CORS
			if o == "*" || o == reqOrigin {
				rw.Header().Set(headerOrigin, reqOrigin)
				allowed = true
				break
			}
		}

		if !allowed {
			webgo.SendHeader(rw, http.StatusForbidden)
			return
		}

		rw.Header().Set(headerMethods, allowMethods)
		rw.Header().Set(headerCreds, "true")

		// Adding allowed headers
		rw.Header().Set(headerAllowHeaders, allowHeaders+req.Header.Get(headerReqHeaders))
	}
}

// CorsOptions is a CORS middleware only for OPTIONS request method
func CorsOptions(allowedOrigins ...string) http.HandlerFunc {
	deprecationLog()
	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		allowed := false
		// Set appropriate response headers required for CORS
		reqOrigin := req.Header.Get(headerGetOrigin)
		for _, o := range allowedOrigins {
			// Set appropriate response headers required for CORS
			if o == "*" || o == reqOrigin {
				rw.Header().Set(headerOrigin, reqOrigin)
				allowed = true
				break
			}
		}

		if !allowed {
			webgo.SendHeader(rw, http.StatusForbidden)
			return
		}

		rw.Header().Set(headerMethods, allowMethods)
		rw.Header().Set(headerCreds, "true")
		rw.Header().Set(headerAllowHeaders, allowHeaders+req.Header.Get(headerReqHeaders))
		webgo.SendHeader(rw, http.StatusOK)
	}
}

// CorsWrap is a single Cors middleware which can be applied to the whole app at once
func CorsWrap(allowedOrigins ...string) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	deprecationLog()
	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}
	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		allowed := false
		// Set appropriate response headers required for CORS
		reqOrigin := req.Header.Get(headerGetOrigin)
		for _, o := range allowedOrigins {
			// Set appropriate response headers required for CORS
			if o == "*" || o == reqOrigin {
				rw.Header().Set(headerOrigin, reqOrigin)
				allowed = true
				break
			}
		}

		if !allowed {
			webgo.SendHeader(rw, http.StatusForbidden)
			return
		}

		rw.Header().Set(headerMethods, allowMethods)
		rw.Header().Set(headerCreds, "true")
		rw.Header().Set(headerAllowHeaders, allowHeaders+req.Header.Get(headerReqHeaders))
		if req.Method == http.MethodOptions {
			webgo.SendHeader(rw, http.StatusOK)
			return
		}

		next(rw, req)
	}
}
