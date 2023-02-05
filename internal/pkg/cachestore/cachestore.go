package cachestore

import (
	"context"
	"fmt"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrCacheMiss is the error returned when the requested item is not available in cache
	ErrCacheMiss = errors.NotFound("not found in cache")
	// ErrCacheNotInitialized is the error returned when the cache handler is not initialized
	ErrCacheNotInitialized = errors.New("not initialized")
)

// Config holds all the configuration required for this package
type Config struct {
	Host string
	Port int

	DB       int
	Username string
	Password string

	PoolSize     int
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

func New(cfg *Config) (*redis.Client, error) {
	cli := redis.NewClient(
		&redis.Options{
			Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Username:        cfg.Username,
			Password:        cfg.Password,
			DB:              cfg.DB,
			DialTimeout:     cfg.DialTimeout,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			ConnMaxIdleTime: cfg.IdleTimeout,
			PoolSize:        cfg.PoolSize,
		},
	)
	err := cli.Ping(context.Background()).Err()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping")
	}
	return cli, nil
}
