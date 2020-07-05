package main

import (
	"log"
	"os"

	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/configs"
	"github.com/bnkamalesh/goapp/internal/platform/datastore"
	"github.com/bnkamalesh/goapp/internal/server/http"
	"github.com/bnkamalesh/goapp/internal/users"
)

func main() {
	l := log.New(os.Stdout, "goapp:", 0)

	cfg, err := configs.NewService()
	if err != nil {
		l.Fatalln(err)
		return
	}

	dscfg, err := cfg.Datastore()
	if err != nil {
		l.Fatalln(err)
		return
	}

	pqdriver, err := datastore.NewService(dscfg)
	if err != nil {
		l.Fatalln(err)
		return
	}

	us, err := users.NewService(pqdriver)
	if err != nil {
		l.Fatalln(err)
		return
	}

	a, err := api.NewService(l, us)
	if err != nil {
		l.Fatalln(err)
		return
	}

	httpCfg, err := cfg.HTTP()
	if err != nil {
		l.Fatalln(err)
		return
	}

	h, err := http.NewService(
		httpCfg,
		a,
	)
	if err != nil {
		l.Fatalln(err)
		return
	}

	h.Start()
}
