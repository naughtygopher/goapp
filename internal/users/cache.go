package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bnkamalesh/errors"
	"github.com/gomodule/redigo/redis"

	"github.com/bnkamalesh/goapp/internal/pkg/cachestore"
)

type userCachestore interface {
	SetUser(ctx context.Context, email string, u *User) error
	ReadUserByEmail(ctx context.Context, email string) (*User, error)
}

type usercache struct {
	pool *redis.Pool
}

func (uc *usercache) conn(ctx context.Context) (redis.Conn, error) {
	return uc.pool.GetContext(ctx)
}

func userCacheKey(id string) string {
	return fmt.Sprintf("user-%s", id)
}

func (uc *usercache) SetUser(ctx context.Context, email string, u *User) error {
	if uc.pool == nil {
		return cachestore.ErrCacheNotInitialized
	}

	conn, err := uc.conn(ctx)
	if err != nil {
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	// it is safe to ignore error here because User struct has no field which can cause the marshal to fail
	payload, _ := json.Marshal(u)

	key := userCacheKey(email)
	_, err = conn.Do("SET", key, payload)
	if err != nil {
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	// expiry in seconds. 1hr
	_, err = conn.Do("EXPIRE", key, 60*60*1)
	if err != nil {
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	return nil
}

func (uc *usercache) ReadUserByEmail(ctx context.Context, email string) (*User, error) {
	if uc.pool == nil {
		return nil, cachestore.ErrCacheNotInitialized
	}

	conn, err := uc.conn(ctx)
	if err != nil {
		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}

	key := userCacheKey(email)

	payload, err := conn.Do("GET", key)
	if err != nil {
		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}
	if payload == nil {
		return nil, cachestore.ErrCacheMiss
	}

	payloadBytes, _ := payload.([]byte)
	u := new(User)
	err = json.Unmarshal(payloadBytes, u)
	if err != nil {
		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}

	return u, nil
}

func newCacheStore(pool *redis.Pool) (*usercache, error) {
	return &usercache{
		pool: pool,
	}, nil
}
