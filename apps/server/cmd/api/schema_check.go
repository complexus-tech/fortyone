package main

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

func runSchemaCheck(ctx context.Context, log *logger.Logger) error {
	cfg, err := parseMigrationProcessConfig()
	if err != nil {
		return err
	}
	mode, err := deployment.Parse(cfg.Environment)
	if err != nil {
		return err
	}
	database := platformdatabase.Config{
		Host: cfg.DB.Host, Port: cfg.DB.Port, User: cfg.DB.User,
		Password: cfg.DB.Password, Name: cfg.DB.Name, MaxOpenConns: 1,
		ConnectTimeout: cfg.DB.ConnectTimeout, SSLMode: cfg.DB.SSLMode,
		SSLRootCert: cfg.DB.SSLRootCert, DisableTLS: cfg.DB.DisableTLS,
	}
	sslMode, err := platformdatabase.EffectiveSSLMode(database)
	if err != nil {
		return err
	}
	if err := deployment.ValidateProductionTransports(mode, deployment.TransportSecurity{PostgreSQLSSLMode: sslMode}); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	schemaVersion, err := migrations.CheckSchema(ctx, platformdatabase.MigrationConfig{Config: database})
	if err != nil {
		return err
	}
	log.Info(ctx, "database schema preflight passed", "schema_version", schemaVersion, "version", version)
	return nil
}
