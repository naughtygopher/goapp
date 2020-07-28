package http

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/webgo/v4"
	"github.com/bnkamalesh/webgo/v4/middleware"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"

	"github.com/bnkamalesh/goapp/internal/api"
)

func errResponder(w http.ResponseWriter, err error) {
	status, msg, _ := errors.HTTPStatusCodeMessage(err)
	webgo.SendError(w, msg, status)
}

// Handlers struct has all the dependencies required for HTTP handlers
type Handlers struct {
	api  *api.API
	home *template.Template
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
		&webgo.Route{
			Name:          "read-user-byemail",
			Pattern:       "/users/:email",
			Method:        http.MethodGet,
			Handlers:      []http.HandlerFunc{h.ReadUserByEmail},
			TrailingSlash: true,
		},
	}
}

// Health is the HTTP handler to return the status of the app including the version, and other details
// This handler uses webgo to respond to the http request
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	out, err := h.api.Health()
	if err != nil {
		errResponder(w, err)
		return
	}
	webgo.R200(w, out)
}

// HelloWorld is a helloworld HTTP handler
func (h *Handlers) HelloWorld(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		{
			webgo.SendResponse(w, "hello world", http.StatusOK)
		}
	default:
		{
			buff := bytes.NewBufferString("")
			err := h.home.Execute(
				buff,
				struct {
					Message string
				}{
					Message: "welcome to the home page!",
				},
			)
			if err != nil {
				webgo.Send(w, contentType, err.Error(), http.StatusInternalServerError)
			}
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.Write(buff.Bytes())
		}
	}
}

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

func panicRecoverer(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		p := recover()
		if p == nil {
			return
		}
		fmt.Println(string(debug.Stack()))
		webgo.R500(w, errors.DefaultMessage)
	}()

	next(w, r)
}

func loadHomeTemplate(basePath string) (*template.Template, error) {
	t := template.New("index.html")
	home, err := t.ParseFiles(
		fmt.Sprintf("%s/index.html", basePath),
	)
	if err != nil {
		return nil, err
	}

	return home, nil
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
		h.routes(),
	)

	router.Use(middleware.AccessLog)
	router.Use(panicRecoverer)
	tracer, _ := apm.NewTracer(
		"goapp",
		"v1.1.3",
	)

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
