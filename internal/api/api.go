package api

import (
	"context"
	"time"

	"github.com/bnkamalesh/goapp/internal/users"
)

var (
	now = time.Now()
)

// Server has all the methods required to run the server
type Server interface {
	CreateUser(ctx context.Context, user *users.User) (*users.User, error)
	ReadUserByEmail(ctx context.Context, email string) (*users.User, error)
	ServerHealth() (map[string]any, error)
}

// Subscriber has all the methods required to run the subscriber
type Subscriber interface {
	AsyncCreateUsers(ctcx context.Context, users []users.User) error
}

type API struct {
	users *users.Users
}

// ServerHealth returns the health of the serever app along with other info like version
func (a *API) ServerHealth() (map[string]any, error) {
	return map[string]any{
		"env":        "testing",
		"version":    "v0.1.0",
		"commit":     "<git commit hash>",
		"status":     "all systems up and running",
		"startedAt":  now.String(),
		"releasedOn": now.String(),
	}, nil

}

func New(us *users.Users) *API {
	return &API{
		users: us,
	}
}

func NewServer(us *users.Users) Server {
	return &API{
		users: us,
	}
}

func NewSubscriber() Subscriber {
	return &API{}
}
