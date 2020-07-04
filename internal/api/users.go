package api

import (
	"context"

	"github.com/bnkamalesh/goapp/internal/users"
)

// CreateUser is the API to create/signup a new user
func (a *API) CreateUser(ctx context.Context, u *users.User) (*users.User, error) {
	u, err := a.users.CreateUser(ctx, u)
	if err != nil {
		a.logger.Println(err)
		return nil, err
	}

	return u, nil
}
