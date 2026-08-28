package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "external-integration ", log.LstdFlags|log.LUTC)
	config, err := loadConfig()
	if err != nil {
		logger.Fatal(err)
	}
	app, err := newIntegration(config, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		if err := app.close(); err != nil {
			logger.Printf("close durable inbox: %v", err)
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/fortyone", app.webhookHandler)
	server := &http.Server{
		Addr:              config.listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 * 1024,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Printf("webhook receiver listening addr=%s", config.listenAddr)
		serverErrors <- server.ListenAndServe()
	}()

	if total, err := app.syncStories(ctx); err != nil {
		logger.Printf("initial story sync failed: %v", err)
	} else {
		logger.Printf("initial story sync complete count=%d", total)
	}
	if config.createStory != nil {
		story, err := app.createStory(ctx, *config.createStory)
		if err != nil {
			logger.Printf("idempotent story create failed: %v", err)
		} else {
			logger.Printf("idempotent story create complete id=%s reference=%s", story.Id, story.Reference)
		}
	}

	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("webhook server stopped: %v", err)
		}
	}
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Printf("webhook server shutdown failed: %v", err)
	}
}
