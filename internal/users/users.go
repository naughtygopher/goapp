package users

import (
	"context"
	"strings"

	"github.com/bnkamalesh/errors"
)

type User struct {
	ID      string
	Name    string
	Address string
	Phone   string
}

// ValidateForCreate runs the validation required for when a user is being created. i.e. ID is not available
func (us *User) ValidateForCreate() error {
	if us.Name == "" {
		return errors.Validation("name cannot be empty")
	}

	if us.Phone == "" {
		return errors.Validation("phone number cannot be empty")
	}

	return nil
}

func (us *User) Sanitize() {
	us.ID = strings.TrimSpace(us.ID)
	us.Name = strings.TrimSpace(us.Name)
	us.Address = strings.TrimSpace(us.Address)
	us.Phone = strings.TrimSpace(us.Phone)
}

type store interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SaveUser(ctx context.Context, user *User) (string, error)
	BulkSaveUser(ctx context.Context, users []User) error
}
type Users struct {
	store store
}

func (us *Users) CreateUser(ctx context.Context, user *User) (*User, error) {
	user.Sanitize()
	err := user.ValidateForCreate()
	if err != nil {
		return nil, err
	}

	newID, err := us.store.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = newID

	return user, nil
}

func (us *Users) ReadByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, errors.Validation("no email provided")
	}

	return us.store.GetUserByEmail(ctx, email)
}

func (us *Users) AsyncCreateUsers(ctx context.Context, users []User) error {
	errList := make([]error, 0, len(users))
	for i := range users {
		err := users[i].ValidateForCreate()
		if err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) != 0 {
		return errors.Join(errList...)
	}

	return us.store.BulkSaveUser(ctx, users)
}
