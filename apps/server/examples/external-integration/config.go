package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

type config struct {
	token         string
	writeToken    string
	workspaceID   uuid.UUID
	webhookSecret string
	baseURL       string
	listenAddr    string
	deliveryLog   string
	createStory   *storyCreateConfig
}

type storyCreateConfig struct {
	teamID         uuid.UUID
	title          string
	idempotencyKey string
}

func loadConfig() (config, error) {
	token := os.Getenv("FORTYONE_TOKEN")
	if token == "" {
		return config{}, errors.New("FORTYONE_TOKEN is required")
	}
	workspaceID, err := uuid.Parse(os.Getenv("FORTYONE_WORKSPACE_ID"))
	if err != nil {
		return config{}, errors.New("FORTYONE_WORKSPACE_ID must be a UUID")
	}
	webhookSecret := os.Getenv("FORTYONE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return config{}, errors.New("FORTYONE_WEBHOOK_SECRET is required")
	}
	listenAddr := strings.TrimSpace(os.Getenv("FORTYONE_LISTEN_ADDR"))
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	deliveryLog := strings.TrimSpace(os.Getenv("FORTYONE_DELIVERY_LOG"))
	if deliveryLog == "" {
		deliveryLog = "./var/webhook-deliveries.log"
	}
	baseURL := strings.TrimSpace(os.Getenv("FORTYONE_API_URL"))
	if strings.HasPrefix(baseURL, "http://") {
		return config{}, fmt.Errorf("FORTYONE_API_URL must use HTTPS: %q", baseURL)
	}
	createStory, err := loadStoryCreateConfig()
	if err != nil {
		return config{}, err
	}
	return config{
		token:         token,
		writeToken:    os.Getenv("FORTYONE_WRITE_TOKEN"),
		workspaceID:   workspaceID,
		webhookSecret: webhookSecret,
		baseURL:       baseURL,
		listenAddr:    listenAddr,
		deliveryLog:   deliveryLog,
		createStory:   createStory,
	}, nil
}

func loadStoryCreateConfig() (*storyCreateConfig, error) {
	team := strings.TrimSpace(os.Getenv("FORTYONE_CREATE_STORY_TEAM_ID"))
	title := strings.TrimSpace(os.Getenv("FORTYONE_CREATE_STORY_TITLE"))
	key := os.Getenv("FORTYONE_CREATE_STORY_IDEMPOTENCY_KEY")
	if team == "" && title == "" && key == "" {
		return nil, nil
	}
	if team == "" || title == "" || key == "" {
		return nil, errors.New("FORTYONE_CREATE_STORY_TEAM_ID, FORTYONE_CREATE_STORY_TITLE, and FORTYONE_CREATE_STORY_IDEMPOTENCY_KEY must be set together")
	}
	teamID, err := uuid.Parse(team)
	if err != nil {
		return nil, errors.New("FORTYONE_CREATE_STORY_TEAM_ID must be a UUID")
	}
	if len(title) > 500 {
		return nil, errors.New("FORTYONE_CREATE_STORY_TITLE must not exceed 500 characters")
	}
	if len(key) < 16 || len(key) > 255 || key != strings.TrimSpace(key) {
		return nil, errors.New("FORTYONE_CREATE_STORY_IDEMPOTENCY_KEY must contain 16 to 255 visible ASCII characters")
	}
	for _, character := range key {
		if character < 0x21 || character > 0x7e {
			return nil, errors.New("FORTYONE_CREATE_STORY_IDEMPOTENCY_KEY must contain 16 to 255 visible ASCII characters")
		}
	}
	return &storyCreateConfig{teamID: teamID, title: title, idempotencyKey: key}, nil
}
