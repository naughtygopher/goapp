package api

import (
	"time"

	"github.com/bnkamalesh/goapp/internal/pkg/logger"
	"github.com/bnkamalesh/goapp/internal/users"
)

var (
	now = time.Now()
)

// API holds all the dependencies required to expose APIs. And each API is a function with *API as its receiver
type API struct {
	logger logger.Logger
	users  *users.Users
}

// Health returns the health of the app along with other info like version
func (a *API) Health() (map[string]interface{}, error) {
	return map[string]interface{}{
		"env":        "testing",
		"version":    "v0.1.0",
		"commit":     "<git commit hash>",
		"status":     "all systems up and running",
		"startedAt":  now.String(),
		"releasedOn": now.String(),
	}, nil

}

// NewService returns a new instance of API with all the dependencies initialized
func NewService(l logger.Logger, us *users.Users) (*API, error) {
	return &API{
		logger: l,
		users:  us,
	}, nil
}
