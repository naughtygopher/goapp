package users

import "context"

type store interface {
	Create(ctx context.Context, u *User) error
	ReadByEmail(ctx context.Context, email string) (*User, error)
}
