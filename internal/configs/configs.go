package configs

import (
	"os"
	"strings"
	"time"

	"github.com/bnkamalesh/goapp/cmd/server/http"
)

type env string

func (e env) String() string {
	return string(e)
}

const (
	EnvLocal      env = "local"
	EnvTest       env = "test"
	EnvStaging    env = "staging"
	EnvProduction env = "production"
)

// Configs struct handles all dependencies required for handling configurations
type Configs struct {
	Environment env
	AppName     string
	AppVersion  string
}

// HTTP returns the configuration required for HTTP package
func (cfg *Configs) HTTP() (*http.Config, error) {
	return &http.Config{
		TemplatesBasePath: strings.TrimSpace(os.Getenv("TEMPLATES_BASEPATH")),
		Port:              8080,
		ReadTimeout:       time.Second * 5,
		WriteTimeout:      time.Second * 5,
		DialTimeout:       time.Second * 3,
	}, nil
}

// New returns an instance of Config with all the required dependencies initialized
func New() (*Configs, error) {
	return &Configs{}, nil
}
