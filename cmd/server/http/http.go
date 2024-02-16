package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/pkg/apm"
	"github.com/bnkamalesh/webgo/v6"
	"github.com/bnkamalesh/webgo/v6/middleware/accesslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Config holds all the configuration required to start the HTTP server
type Config struct {
	Host string
	Port uint16

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration

	TemplatesBasePath string
}

type HTTP struct {
	apis     api.Server
	listener string
	server   *http.Server
}

// Start starts the HTTP server
func (h *HTTP) Start() error {
	webgo.LOGHANDLER.Info("HTTP server, listening on", h.listener)
	return h.server.ListenAndServe()
}

// NewService returns an instance of HTTP with all its dependencies set
func NewService(cfg *Config, apis api.Server) (*HTTP, error) {
	home, err := loadHomeTemplate(cfg.TemplatesBasePath)
	if err != nil {
		return nil, err
	}

	handlers := &Handlers{
		apis: apis,
		home: home,
	}

	router := webgo.NewRouter(
		&webgo.Config{
			Host:            cfg.Host,
			Port:            strconv.Itoa(int(cfg.Port)),
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			ShutdownTimeout: cfg.WriteTimeout * 2,
		},
		handlers.routes()...,
	)

	router.Use(accesslog.AccessLog)
	router.Use(panicRecoverer)

	// in this app, /-/ prefixed routes are used for healthchecks, readiness checks etc.
	_ = otelhttp.WithFilter(func(req *http.Request) bool {
		return !strings.HasPrefix(req.URL.Path, "/-/")
	})

	// the span name formatter is used to reduce the cardinality of metrics generated
	// when using URIs with variables in it
	otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
		wctx := webgo.Context(r)
		if wctx == nil {
			return r.URL.Path
		}
		return wctx.Route.Pattern
	})

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           apm.NewHTTPMiddleware()(router),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.ReadTimeout * 2,
	}

	return &HTTP{
		server:   httpServer,
		listener: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}, nil
}
