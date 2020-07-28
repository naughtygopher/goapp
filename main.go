package main

import (
	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/configs"
	"github.com/bnkamalesh/goapp/internal/platform/cachestore"
	"github.com/bnkamalesh/goapp/internal/platform/datastore"
	"github.com/bnkamalesh/goapp/internal/platform/logger"
	"github.com/bnkamalesh/goapp/internal/server/http"
	"github.com/bnkamalesh/goapp/internal/users"
)

func main() {
	l := logger.New("goapp", "v1.0.0", 1)

	cfg, err := configs.NewService()
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	dscfg, err := cfg.Datastore()
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	pqdriver, err := datastore.NewService(dscfg)
	if err != nil {
		// l.Fatal(err.Error())
		// return
	}

	cacheCfg, err := cfg.Cachestore()
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	redispool, err := cachestore.NewService(cacheCfg)
	if err != nil {
		// Cache could be something we'd be willing to tolerate if not available
		// Though this is strictly based on how critical cache is to your application
		l.Error(err)
	}

	us, err := users.NewService(l, pqdriver, redispool)
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	a, err := api.NewService(l, us)
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	httpCfg, err := cfg.HTTP()
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	h, err := http.NewService(
		httpCfg,
		a,
	)
	if err != nil {
		l.Fatal(err.Error())
		return
	}

	h.Start()
}
