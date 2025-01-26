package api

import (
	"context"

	"github.com/naughtygopher/goapp/internal/usernotes"
	"github.com/naughtygopher/goapp/internal/users"
)

// Server has all the methods required to run the server
type Server interface {
	CreateUser(ctx context.Context, user *users.User) (*users.User, error)
	ReadUserByEmail(ctx context.Context, email string) (*users.User, error)
	CreateUserNote(ctx context.Context, un *usernotes.Note) (*usernotes.Note, error)
	ReadUserNote(ctx context.Context, userID string, noteID string) (*usernotes.Note, error)
}

// Subscriber has all the methods required to run the subscriber
type Subscriber interface {
	AsyncCreateUsers(ctcx context.Context, users []users.User) error
}

type API struct {
	users  *users.Users
	unotes *usernotes.UserNotes
}

func New(us *users.Users, un *usernotes.UserNotes) *API {
	return &API{
		users:  us,
		unotes: un,
	}
}

func NewServer(us *users.Users, un *usernotes.UserNotes) Server {
	return New(us, un)
}

func NewSubscriber(us *users.Users) Subscriber {
	return New(us, nil)
}
