package cachestore

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/gomodule/redigo/redis"
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
	Port string

	StoreName string
	Username  string
	Password  string

	PoolSize     int
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

// NewService returns an instance of redis.Pool with all the required configurations set
func NewService(cfg *Config) (*redis.Pool, error) {
	db, _ := strconv.Atoi(cfg.StoreName)
	rpool := &redis.Pool{
		MaxIdle:         cfg.PoolSize,
		MaxActive:       cfg.PoolSize,
		IdleTimeout:     cfg.IdleTimeout,
		Wait:            true,
		MaxConnLifetime: cfg.IdleTimeout * 2,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
				redis.DialReadTimeout(cfg.ReadTimeout),
				redis.DialWriteTimeout(cfg.WriteTimeout),
				redis.DialPassword(cfg.Password),
				redis.DialConnectTimeout(cfg.DialTimeout),
				redis.DialDatabase(db),
			)
		},
	}

	conn := rpool.Get()
	rep, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}

	pong, _ := rep.(string)
	if pong != "PONG" {
		return nil, errors.New("ping failed")
	}
	conn.Close()

	return rpool, nil
}
