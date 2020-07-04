package main

import (
	"log"
	"os"
	"time"

	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/server/http"
	"github.com/bnkamalesh/goapp/internal/users"
)

func main() {
	l := log.New(os.Stdout, "goapp", 0)
	us, err := users.NewService()
	if err != nil {
		l.Fatalln(err)
		return
	}

	a, err := api.NewService(l, us)
	if err != nil {
		l.Fatalln(err)
		return
	}

	h, err := http.NewService(
		&http.Config{
			Port:         "8080",
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
			DialTimeout:  time.Second * 3,
		},
		a,
	)
	if err != nil {
		l.Fatalln(err)
		return
	}

	h.Start()
}
