package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/naughtygopher/errors"
	"github.com/naughtygopher/proberesponder"
	proberespHTTP "github.com/naughtygopher/proberesponder/extensions/http"

	"github.com/naughtygopher/goapp/cmd/server/grpc"
	xhttp "github.com/naughtygopher/goapp/cmd/server/http"
	"github.com/naughtygopher/goapp/internal/api"
	"github.com/naughtygopher/goapp/internal/configs"
	"github.com/naughtygopher/goapp/internal/pkg/apm"
	"github.com/naughtygopher/goapp/internal/pkg/logger"
	"github.com/naughtygopher/goapp/internal/pkg/postgres"
	"github.com/naughtygopher/goapp/internal/users"
)

func startAPM(ctx context.Context, cfg *configs.Configs) *apm.APM {
	ap, err := apm.New(ctx, &apm.Options{
		Debug:            cfg.Environment == configs.EnvLocal,
		Environment:      cfg.Environment.String(),
		ServiceName:      cfg.AppName,
		ServiceVersion:   cfg.AppVersion,
		TracesSampleRate: 50.00,
		UseStdOut:        cfg.Environment == configs.EnvLocal,
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
		err = hserver.Start()
		if err != nil {
			fatalErr <- errors.Wrap(err, "failed to start HTTP server")
		}
	}()

	return hserver, nil
}

func startHealthResponder(ctx context.Context, ps *proberesponder.ProbeResponder, api *api.API, fatalErr chan<- error) (*http.Server, error) {
	port := uint32(2000)
	srv := proberespHTTP.Server(ps, "", uint16(port), proberespHTTP.Handlers{http.MethodGet, "/-/health", func(w http.ResponseWriter, r *http.Request) {}})
	go func() {
		defer logger.Info(ctx, fmt.Sprintf("[http/healthresponder] :%d shutdown complete", port))
		logger.Info(ctx, fmt.Sprintf("[http/healthresponder] listening on :%d", port))
		fatalErr <- srv.ListenAndServe()
	}()
	return srv, nil
}

func start(
	ctx context.Context,
	cfgs *configs.Configs,
	fatalErr chan<- error,
) (hserver *xhttp.HTTP, gserver *grpc.GRPC) {
	_ = ctx
	pqdriver, err := postgres.NewPool(cfgs.Postgres())
	if err != nil {
		panic(errors.Wrap(err))
	}

	userPGstore := users.NewPostgresStore(pqdriver, cfgs.UserPostgresTable())
	userSvc := users.NewService(userPGstore)
	svrAPIs := api.NewServer(userSvc, nil)
	hserver, gserver = startServers(svrAPIs, cfgs, fatalErr)
	return
}
