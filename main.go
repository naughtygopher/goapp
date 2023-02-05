package main

import (
	"fmt"

	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/configs"
	"github.com/bnkamalesh/goapp/internal/pkg/cachestore"
	"github.com/bnkamalesh/goapp/internal/pkg/datastore"
	"github.com/bnkamalesh/goapp/internal/pkg/logger"
	"github.com/bnkamalesh/goapp/internal/server/http"
	"github.com/bnkamalesh/goapp/internal/users"
)

func main() {
	l := logger.New("goapp", "v1.0.0", 1)
	logger.UpdateDefaultLogger(l)

	cfg, err := configs.New()
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	dscfg, err := cfg.Datastore()
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	pqdriver, err := datastore.NewService(dscfg)
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	cacheCfg, err := cfg.Cachestore()
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	redispool, err := cachestore.New(cacheCfg)
	if err != nil {
		// Cache could be something we'd be willing to tolerate if not available.
		// Though this is strictly based on how critical cache is to your application
		_ = logger.Error(fmt.Sprintf("%+v", err))
	}

	userStore, err := users.NewStore(pqdriver)
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	userCache, err := users.NewCacheStore(redispool)
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	us, err := users.NewService(l, userStore, userCache)
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	a, err := api.NewService(l, us)
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	httpCfg, err := cfg.HTTP()
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	h, err := http.NewService(
		httpCfg,
		a,
	)
	if err != nil {
		_ = logger.Fatal(fmt.Sprintf("%+v", err))
		return
	}

	_ = logger.Fatal(h.Start())
}
