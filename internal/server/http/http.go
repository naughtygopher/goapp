package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bnkamalesh/webgo/v6"
	"github.com/bnkamalesh/webgo/v6/middleware/accesslog"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"

	"github.com/bnkamalesh/goapp/internal/api"
)

// HTTP struct holds all the dependencies required for starting HTTP server
type HTTP struct {
	server *http.Server
	cfg    *Config
}

// Start starts the HTTP server
func (h *HTTP) Start() {
	webgo.LOGHANDLER.Info("HTTP server, listening on", h.cfg.Host, h.cfg.Port)
	h.server.ListenAndServe()
}

// Config holds all the configuration required to start the HTTP server
type Config struct {
	Host              string
	Port              string
	TemplatesBasePath string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	DialTimeout       time.Duration
}

// NewService returns an instance of HTTP with all its dependencies set
func NewService(cfg *Config, a *api.API) (*HTTP, error) {
	home, err := loadHomeTemplate(cfg.TemplatesBasePath)
	if err != nil {
		return nil, err
	}

	h := &Handlers{
		api:  a,
		home: home,
	}

	router := webgo.NewRouter(
		&webgo.Config{
			Host:            cfg.Host,
			Port:            cfg.Port,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			ShutdownTimeout: cfg.WriteTimeout * 2,
		},
		h.routes()...,
	)

	router.Use(accesslog.AccessLog)
	router.Use(panicRecoverer)
	tracer, _ := apm.NewTracer("goapp", "v1.1.3")

	serverHandler := apmhttp.Wrap(
		router,
		apmhttp.WithRecovery(apmhttp.NewTraceRecovery(
			tracer,
		)),
	)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:           serverHandler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.ReadTimeout * 2,
	}

	return &HTTP{
		server: httpServer,
		cfg:    cfg,
	}, nil
}
