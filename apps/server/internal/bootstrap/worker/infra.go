package workerbootstrap

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

func openDB(cfg Config) (*sqlx.DB, error) {
	db, err := database.Open(database.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return db, nil
}

func redisClientOpt(cfg Config) asynq.RedisClientOpt {
	var tlsConfig *tls.Config
	if !cfg.Redis.DisableTLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.Redis.InsecureSkipVerify,
		}
	}

	return asynq.RedisClientOpt{
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
