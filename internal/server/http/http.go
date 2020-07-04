package http

import (
	"net/http"
	"time"

	"github.com/bnkamalesh/webgo/v4"
	"github.com/bnkamalesh/webgo/v4/middleware"

	"github.com/bnkamalesh/goapp/internal/api"
)

// Handlers struct has all the dependencies required for HTTP handlers
type Handlers struct {
	api *api.API
}

func (h *Handlers) routes() []*webgo.Route {
	return []*webgo.Route{
		&webgo.Route{
			Name:          "helloworld",
			Pattern:       "",
			Method:        http.MethodGet,
			Handlers:      []http.HandlerFunc{h.HelloWorld},
			TrailingSlash: true,
		},
		&webgo.Route{
			Name:          "health",
			Pattern:       "/-/health",
			Method:        http.MethodGet,
			Handlers:      []http.HandlerFunc{h.Health},
			TrailingSlash: true,
		},
		&webgo.Route{
			Name:          "create-user",
			Pattern:       "/users",
			Method:        http.MethodPost,
			Handlers:      []http.HandlerFunc{h.CreateUser},
			TrailingSlash: true,
		},
	}
}

// Health is the HTTP handler to return the status of the app including the version, and other details
// This handler uses webgo to respond to the http request
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	out, err := h.api.Health()
	if err != nil {
		webgo.R500(w, err.Error())
		return
	}
	webgo.R200(w, out)
}

// HelloWorld is a helloworld HTTP handler
func (h *Handlers) HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// HTTP struct holds all the dependencies required for starting HTTP server
type HTTP struct {
	router *webgo.Router
}

// Start starts the HTTP server
func (h *HTTP) Start() {
	h.router.Use(middleware.AccessLog)
	h.router.Start()
}

// Config holds all the configuration required to start the HTTP server
type Config struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

// NewService returns an instance of HTTP with all its dependencies set
func NewService(cfg *Config, a *api.API) (*HTTP, error) {
	h := &Handlers{
		api: a,
	}

	router := webgo.NewRouter(
		&webgo.Config{
			Host:            cfg.Host,
			Port:            cfg.Port,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			ShutdownTimeout: cfg.WriteTimeout * 2,
		},
		h.routes(),
	)

	return &HTTP{
		router: router,
	}, nil
}
