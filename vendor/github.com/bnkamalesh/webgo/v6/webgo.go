/*
Package webgo is a lightweight framework for building web apps. It has a multiplexer,
middleware plugging mechanism & context management of its own. The primary goal
of webgo is to get out of the developer's way as much as possible. i.e. it does
not enforce you to build your app in any particular pattern, instead just helps you
get all the trivial things done faster and easier.

e.g.
1. Getting named URI parameters.
2. Multiplexer for regex matching of URI and such.
3. Inject special app level configurations or any such objects to the request context as required.
*/
package webgo

import (
	"context"
	"crypto/tls"
	"net/http"
)

var supportedHTTPMethods = []string{
	http.MethodOptions,
	http.MethodHead,
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// ctxkey is a custom string type to store the WebGo context inside HTTP request context
type ctxkey string

const wgoCtxKey = ctxkey("webgocontext")

// ContextPayload is the WebgoContext. A new instance of ContextPayload is injected inside every request's context object
type ContextPayload struct {
	Route *Route
	Err   error
	path  string
}

// Params returns the URI parameters of the respective route
func (cp *ContextPayload) Params() map[string]string {
	return cp.Route.params(cp.path)
}

func (cp *ContextPayload) reset() {
	cp.Route = nil
	cp.path = ""
	cp.Err = nil
}

// SetError sets the err within the context
func (cp *ContextPayload) SetError(err error) {
	cp.Err = err
}

// Error returns the error set within the context
func (cp *ContextPayload) Error() error {
	return cp.Err
}

// Context returns the ContextPayload injected inside the HTTP request context
func Context(r *http.Request) *ContextPayload {
	return r.Context().Value(wgoCtxKey).(*ContextPayload)
}

// SetError is a helper function to set the error in webgo context
func SetError(r *http.Request, err error) {
	ctx := Context(r)
	ctx.SetError(err)
}

// GetError is a helper function to get the error from webgo context
func GetError(r *http.Request) error {
	return Context(r).Error()
}

// ResponseStatus returns the response status code. It works only if the http.ResponseWriter
// is not wrapped in another response writer before calling ResponseStatus
func ResponseStatus(rw http.ResponseWriter) int {
	crw, ok := rw.(*customResponseWriter)
	if !ok {
		return http.StatusOK
	}
	return crw.statusCode
}

// StartHTTPS starts the server with HTTPS enabled
func (router *Router) StartHTTPS() {
	cfg := router.config
	if cfg.CertFile == "" {
		LOGHANDLER.Fatal("No certificate provided for HTTPS")
	}

	if cfg.KeyFile == "" {
		LOGHANDLER.Fatal("No key file provided for HTTPS")
	}

	host := cfg.Host
	if len(cfg.HTTPSPort) > 0 {
		host += ":" + cfg.HTTPSPort
	}

	router.httpsServer = &http.Server{
		Addr:         host,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	LOGHANDLER.Info("HTTPS server, listening on", host)
	err := router.httpsServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	if err != nil && err != http.ErrServerClosed {
		LOGHANDLER.Error("HTTPS server exited with error:", err.Error())
	}
}

// Start starts the HTTP server with the appropriate configurations
func (router *Router) Start() {
	cfg := router.config
	host := cfg.Host

	if len(cfg.Port) > 0 {
		host += ":" + cfg.Port
	}

	router.httpServer = &http.Server{
		Addr:         host,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	LOGHANDLER.Info("HTTP server, listening on", host)
	err := router.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		LOGHANDLER.Error("HTTP server exited with error:", err.Error())
	}
}

// Shutdown gracefully shuts down HTTP server
func (router *Router) Shutdown() error {
	if router.httpServer == nil {
		return nil
	}
	timer := router.config.ShutdownTimeout

	ctx, cancel := context.WithTimeout(context.TODO(), timer)
	defer cancel()

	err := router.httpServer.Shutdown(ctx)
	if err != nil {
		LOGHANDLER.Error(err)
	}
	return err
}

// ShutdownHTTPS gracefully shuts down HTTPS server
func (router *Router) ShutdownHTTPS() error {
	if router.httpsServer == nil {
		return nil
	}
	timer := router.config.ShutdownTimeout

	ctx, cancel := context.WithTimeout(context.TODO(), timer)
	defer cancel()

	err := router.httpsServer.Shutdown(ctx)
	if err != nil && err != http.ErrServerClosed {
		LOGHANDLER.Error(err)
	}
	return err
}

// OriginalResponseWriter returns the Go response writer stored within the webgo custom response
// writer
func OriginalResponseWriter(rw http.ResponseWriter) http.ResponseWriter {
	crw, ok := rw.(*customResponseWriter)
	if !ok {
		return nil
	}

	return crw.ResponseWriter
}
