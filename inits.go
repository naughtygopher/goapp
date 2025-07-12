package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/naughtygopher/errors"
	"github.com/naughtygopher/proberesponder"
	"github.com/naughtygopher/proberesponder/extensions/depprober"
	proberespHTTP "github.com/naughtygopher/proberesponder/extensions/http"
	"github.com/naughtygopher/webgo/v7"

	"github.com/naughtygopher/goapp/cmd/server/grpc"
	xhttp "github.com/naughtygopher/goapp/cmd/server/http"
	"github.com/naughtygopher/goapp/internal/api"
	"github.com/naughtygopher/goapp/internal/configs"
	"github.com/naughtygopher/goapp/internal/pkg/apm"
	"github.com/naughtygopher/goapp/internal/pkg/logger"
	"github.com/naughtygopher/goapp/internal/pkg/postgres"
	"github.com/naughtygopher/goapp/internal/users"
)

var now = time.Now()

func startAPM(ctx context.Context, cfg *configs.Configs) *apm.APM {
	ap, err := apm.New(ctx, &apm.Options{
		Debug:                cfg.Environment == configs.EnvLocal,
		Environment:          cfg.Environment.String(),
		ServiceName:          cfg.AppName,
		ServiceVersion:       cfg.AppVersion,
		PrometheusScrapePort: 9090,
		TracesSampleRate:     50.00,
		UseStdOut:            cfg.Environment == configs.EnvLocal,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to start APM"))
	}
	return ap
}

func startServers(svr api.Server, cfgs *configs.Configs, fatalErr chan<- error) (*xhttp.HTTP, *grpc.GRPC) {
	hcfg, _ := cfgs.HTTP()
	hserver, err := xhttp.NewService(hcfg, svr)
	if err != nil {
		fatalErr <- errors.Wrap(err, "failed to initialize HTTP server")
	}

	go func() {
		defer func() {
			rec := recover()
			if rec != nil {
				fatalErr <- errors.New(fmt.Sprintf("%+v", rec))
			}
		}()
		err = hserver.Start()
		if err != nil {
			fatalErr <- errors.Wrap(err, "failed to start HTTP server")
		}
	}()

	return hserver, nil
}

func healthResponseHandler(ps *proberesponder.ProbeResponder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{
			"env":        "testing",
			"version":    "v0.1.0",
			"commit":     "<git commit hash>",
			"status":     "all systems up and running",
			"startedAt":  now.String(),
			"releasedOn": now.String(),
		}

		for key, value := range ps.HealthResponse() {
			payload[key] = value
		}
		b, _ := json.Marshal(payload)
		w.Header().Add(webgo.HeaderContentType, webgo.JSONContentType)
		_, _ = w.Write(b)
	}
}

func startHealthResponder(ctx context.Context, ps *proberesponder.ProbeResponder, fatalErr chan<- error) (*http.Server, error) {
	port := uint32(2000)
	srv := proberespHTTP.Server(
		ps, "", uint16(port),
		proberespHTTP.Handler{
			Method:  http.MethodGet,
			Path:    "/-/health",
			Handler: healthResponseHandler(ps),
		},
	)

	go func() {
		defer logger.Info(ctx, fmt.Sprintf("[http/healthresponder] :%d shutdown complete", port))
		logger.Info(ctx, fmt.Sprintf("[http/healthresponder] listening on :%d", port))
		fatalErr <- srv.ListenAndServe()
	}()

	return srv, nil
}

func start(
	ctx context.Context,
	probestatus *proberesponder.ProbeResponder,
	cfgs *configs.Configs,
	fatalErr chan<- error,
) (hserver *xhttp.HTTP, gserver *grpc.GRPC) {
	_ = ctx
	pqdriver, err := postgres.NewPool(cfgs.Postgres())
	if err != nil {
		panic(errors.Wrap(err))
	}

	depprober.Start(time.Minute, probestatus, &depprober.Probe{
		ID:               "postgres",
		AffectedStatuses: []proberesponder.Statuskey{proberesponder.StatusLive, proberesponder.StatusReady},
		Checker: depprober.CheckerFunc(func(ctx context.Context) error {
			err := pqdriver.Ping(ctx)
			if err != nil {
				return errors.Wrap(err, "postgres ping failed")
			}
			return nil
		}),
	})

	userPGstore := users.NewPostgresStore(pqdriver, cfgs.UserPostgresTable())
	userSvc := users.NewService(userPGstore)
	svrAPIs := api.NewServer(userSvc, nil)
	hserver, gserver = startServers(svrAPIs, cfgs, fatalErr)
	return
}
