package postgres

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/naughtygopher/errors"

	"github.com/naughtygopher/goapp/internal/pkg/apm"
)

const (
	minConn           = 2
	maxConnLifetime   = time.Minute * 5
	healthCheckPeriod = time.Second * 30
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

func pgxPoolConfig(cfg *Config) (*pgxpool.Config, error) {
	uri := cfg.ConnURL()
	pgxconfig, err := pgx.ParseConfig(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing connection string")
	}

	pgxconfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithTracerProvider(apm.Global().GetTracerProvider()),
	)

	poolconfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing connection string for pool")
	}

	poolconfig.ConnConfig = pgxconfig

	// poolconfig.BeforeConnect func(context.Context, *pgx.ConnConfig) error
	// poolconfig.AfterConnect func(context.Context, *pgx.Conn) error
	// poolconfig.BeforeAcquire func(context.Context, *pgx.Conn) bool
	// poolconfig.AfterRelease func(*pgx.Conn) bool
	// poolconfig.BeforeClose func(*pgx.Conn)

	// MaxConnLifetime is the duration since creation after which a connection will be automatically closed.
	poolconfig.MaxConnLifetime = maxConnLifetime
	// MaxConnLifetimeJitter is the duration after MaxConnLifetime to randomly decide to close a connection.
	// This helps prevent all connections from being closed at the exact same time, starving the pool.
	poolconfig.MaxConnLifetimeJitter = time.Hour
	// MaxConnIdleTime is the duration after which an idle connection will be automatically closed by the health check.
	poolconfig.MaxConnIdleTime = time.Microsecond
	// MaxConns is the maximum size of the pool. The default is the greater of 4 or runtime.NumCPU().
	// MaxConns:
	poolconfig.MinConns = minConn
	// HealthCheckPeriod is the duration between checks of the health of idle connections.
	poolconfig.HealthCheckPeriod = healthCheckPeriod
	return poolconfig, nil
}

// NewPool returns a new instance of PGX pool
func NewPool(cfg *Config) (*pgxpool.Pool, error) {
	poolcfg, err := pgxPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	poolcfg.MaxConnLifetime = cfg.IdleTimeout
	poolcfg.MaxConns = int32(cfg.ConnPoolSize)

	dialer := &net.Dialer{KeepAlive: cfg.DialTimeout}
	dialer.Timeout = cfg.DialTimeout
	poolcfg.ConnConfig.DialFunc = dialer.DialContext

	pool, err := pgxpool.NewWithConfig(context.Background(), poolcfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pgx pool")
	}

	return pool, nil
}
