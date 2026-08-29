package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"time"

	fortyone "github.com/complexus-tech/fortyone-go"
	"github.com/google/uuid"
)

const maxWebhookBody = 256 * 1024

type integration struct {
	client      *fortyone.ClientWithResponses
	writeClient *fortyone.ClientWithResponses
	workspaceID uuid.UUID
	verifier    *fortyone.WebhookVerifier
	deliveries  *deliveryStore
	logger      *log.Logger
}

type webhookEnvelope struct {
	ID             uuid.UUID       `json:"id"`
	Type           string          `json:"type"`
	PayloadVersion int             `json:"payload_version"`
	OccurredAt     time.Time       `json:"occurred_at"`
	Data           json.RawMessage `json:"data"`
}

func (app *integration) syncStories(ctx context.Context) (int, error) {
	pager, err := fortyone.NewStoryPager(app.client, app.workspaceID, fortyone.StoryPaginationOptions{Limit: 100})
	if err != nil {
		return 0, err
	}
	total := 0
	for {
		page, ok, err := pager.NextPage(ctx)
		if err != nil {
			return total, err
		}
		if !ok {
			return total, nil
		}
		for _, story := range page.Data {
			app.logger.Printf("story visible id=%s reference=%s", story.Id, story.Reference)
			total++
		}
	}
}

func (app *integration) createStory(ctx context.Context, input storyCreateConfig) (fortyone.ComponentsResourcesStory, error) {
	client := app.writeClient
	if client == nil {
		client = app.client
	}
	if client == nil {
		return fortyone.ComponentsResourcesStory{}, errors.New("story write client is not configured")
	}
	response, err := client.CreateStoryWithResponse(
		ctx,
		app.workspaceID,
		&fortyone.CreateStoryParams{IdempotencyKey: input.idempotencyKey},
		fortyone.ComponentsResourcesCreateStoryRequest{Title: input.title, TeamId: input.teamID},
	)
	if err != nil {
		return fortyone.ComponentsResourcesStory{}, fmt.Errorf("create story request: %w", err)
	}
	if response.JSON201 != nil {
		return response.JSON201.Data, nil
	}
	if response.HTTPResponse == nil {
		return fortyone.ComponentsResourcesStory{}, errors.New("create story returned no HTTP response")
	}
	return fortyone.ComponentsResourcesStory{}, fortyone.NewAPIError(
		response.StatusCode(),
		response.HTTPResponse.Header,
		response.Body,
	)
}

func (app *integration) webhookHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(writer, "application/json is required", http.StatusUnsupportedMediaType)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxWebhookBody)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(writer, "request body is too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(writer, "could not read request", http.StatusBadRequest)
		return
	}
	verified, err := app.verifier.Verify(body, request.Header)
	if err != nil {
		http.Error(writer, "webhook verification failed", http.StatusUnauthorized)
		return
	}
	var envelope webhookEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.ID == uuid.Nil || envelope.Type == "" || envelope.PayloadVersion < 1 || envelope.OccurredAt.IsZero() || len(envelope.Data) == 0 {
		http.Error(writer, "webhook envelope is invalid", http.StatusBadRequest)
		return
	}
	alreadySeen, err := app.deliveries.record(verified.ID, body)
	if err != nil {
		app.logger.Printf("record webhook failed id=%s", verified.ID)
		http.Error(writer, "webhook could not be recorded", http.StatusServiceUnavailable)
		return
	}
	if !alreadySeen {
		app.logger.Printf("webhook accepted delivery_id=%s event_id=%s type=%s payload_version=%d", verified.ID, envelope.ID, envelope.Type, envelope.PayloadVersion)
	}
	writer.WriteHeader(http.StatusNoContent)
}

func newIntegration(config config, logger *log.Logger) (*integration, error) {
	client, err := fortyone.New(fortyone.Config{Token: config.token, BaseURL: config.baseURL})
	if err != nil {
		return nil, fmt.Errorf("create FortyOne client: %w", err)
	}
	writeClient := client
	if config.writeToken != "" {
		writeClient, err = fortyone.New(fortyone.Config{Token: config.writeToken, BaseURL: config.baseURL})
		if err != nil {
			return nil, fmt.Errorf("create FortyOne service-account client: %w", err)
		}
	}
	verifier, err := fortyone.NewWebhookVerifier(config.webhookSecret, 0)
	if err != nil {
		return nil, fmt.Errorf("create webhook verifier: %w", err)
	}
	deliveries, err := openDeliveryStore(config.deliveryLog)
	if err != nil {
		return nil, err
	}
	return &integration{
		client: client, writeClient: writeClient, workspaceID: config.workspaceID,
		verifier: verifier, deliveries: deliveries, logger: logger,
	}, nil
}

func (app *integration) close() error { return app.deliveries.close() }
