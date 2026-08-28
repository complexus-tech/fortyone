package workerbootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func openDB(ctx context.Context, cfg Config) (*platformdatabase.Connections, error) {
	connections, err := platformdatabase.Open(ctx, platformdatabase.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		MinConns:     cfg.DB.MinConns,

		ConnectTimeout:    cfg.DB.ConnectTimeout,
		MaxConnIdleTime:   cfg.DB.MaxConnIdleTime,
		MaxConnLifetime:   cfg.DB.MaxConnLifetime,
		HealthCheckPeriod: cfg.DB.HealthCheckPeriod,
		SSLMode:           cfg.DB.SSLMode,
		SSLRootCert:       cfg.DB.SSLRootCert,
		DisableTLS:        cfg.DB.DisableTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return connections, nil
}

func redisClientOpt(cfg Config) asynq.RedisClientOpt {
	options := redisOptions(cfg)

	return asynq.RedisClientOpt{
		Addr:         options.Addr,
		Password:     options.Password,
		DB:           options.DB,
		TLSConfig:    options.TLSConfig,
		DialTimeout:  options.DialTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		PoolSize:     options.PoolSize,
	}
}

func openRedis(cfg Config) *redis.Client {
	return redis.NewClient(redisOptions(cfg))
}

func redisOptions(cfg Config) *redis.Options {
	var tlsConfig *tls.Config
	if !cfg.Redis.DisableTLS {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: cfg.Redis.Host,
		}
	}

	return &redis.Options{
		Addr:         net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Name,
		TLSConfig:    tlsConfig,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
	}
}
