package users

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/redis/go-redis/v9"

	"github.com/bnkamalesh/goapp/internal/pkg/cachestore"
)

type usercache struct {
	cli *redis.Client
}

func (uc *usercache) conn(ctx context.Context) (*redis.Conn, error) {
	return uc.cli.Conn(), nil
}

func userCacheKey(id string) string {
	return fmt.Sprintf("user-%s", id)
}

func (uc *usercache) SetUser(ctx context.Context, email string, u *User) error {
	if uc.cli == nil {
		return cachestore.ErrCacheNotInitialized
	}

	conn, err := uc.conn(ctx)
	if err != nil {
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	// it is safe to ignore error here because User struct has no field which can cause the marshal to fail
	payload, _ := json.Marshal(u)

	key := userCacheKey(email)

	err = conn.Set(ctx, key, payload, time.Hour).Err()
	if err != nil {
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	return nil
}

func (uc *usercache) ReadUserByEmail(ctx context.Context, email string) (*User, error) {
	if uc.cli == nil {
		return nil, cachestore.ErrCacheNotInitialized
	}

	conn, err := uc.conn(ctx)
	if err != nil {
		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}

	key := userCacheKey(email)

	cmd := conn.Get(ctx, key)
	payload, err := cmd.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, errors.DefaultMessage)
	}
	if payload == nil {
		return nil, cachestore.ErrCacheMiss
	}

	u := new(User)
	err = json.Unmarshal(payload, u)
	if err != nil {
		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}

	return u, nil
}

func NewCacheStore(cli *redis.Client) (*usercache, error) {
	return &usercache{
		cli: cli,
	}, nil
}
