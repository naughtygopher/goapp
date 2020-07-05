package datastore

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Config struct holds all the configurations required the datastore package
type Config struct {
	Host   string `json:"host,omitempty"`
	Port   string `json:"port,omitempty"`
	Driver string `json:"driver,omitempty"`

	StoreName string `json:"storeName,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`

	SSLMode string `json:"sslMode,omitempty"`

	ConnPoolSize uint          `json:"connPoolSize,omitempty"`
	ReadTimeout  time.Duration `json:"readTimeout,omitempty"`
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"`
	IdleTimeout  time.Duration `json:"idleTimeout,omitempty"`
	DialTimeout  time.Duration `json:"dialTimeout,omitempty"`
}

// ConnURL returns the connection URL
func (cfg *Config) ConnURL() string {
	sslMode := strings.TrimSpace(cfg.SSLMode)
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Driver,
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.StoreName,
		sslMode,
	)
}

// NewService returns a new instance of PGX pool
func NewService(cfg *Config) (*pgxpool.Pool, error) {
	poolcfg, err := pgxpool.ParseConfig(cfg.ConnURL())
	if err != nil {
		return nil, err
	}

	poolcfg.MaxConnLifetime = cfg.IdleTimeout
	poolcfg.MaxConns = int32(cfg.ConnPoolSize)

	dialer := &net.Dialer{KeepAlive: cfg.DialTimeout}
	dialer.Timeout = cfg.DialTimeout
	poolcfg.ConnConfig.DialFunc = dialer.DialContext

	pool, err := pgxpool.ConnectConfig(context.Background(), poolcfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
