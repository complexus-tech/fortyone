package mailerlite

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/mailerlite/mailerlite-go"
)

// Service wraps the MailerLite client with additional functionality
type Service struct {
	client *mailerlite.Client
	log    *logger.Logger
}

// Config holds the configuration for MailerLite service
type Config struct {
	APIKey string
}

// NewService creates a new MailerLite service instance
func NewService(log *logger.Logger, config Config) (*Service, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := mailerlite.NewClient(config.APIKey)

	return &Service{
		client: client,
		log:    log,
	}, nil
}

// CreateOrUpdateSubscriber creates/updates a subscriber and returns the subscriber ID
func (s *Service) CreateOrUpdateSubscriber(ctx context.Context, email, fullName string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("MailerLite client not initialized")
	}

	// Create/Upsert subscriber
	subscriber := &mailerlite.UpsertSubscriber{
		Email: email,
		Fields: map[string]any{
			"name": fullName,
		},
	}

	s.log.Info(ctx, "Creating MailerLite subscriber", "email", email, "name", fullName)

	newSubscriber, _, err := s.client.Subscriber.Upsert(ctx, subscriber)
	if err != nil {
		s.log.Error(ctx, "Failed to create/update MailerLite subscriber", "error", err, "email", email)
		return "", fmt.Errorf("failed to create/update MailerLite subscriber: %w", err)
	}

	s.log.Info(ctx, "MailerLite subscriber created/updated",
		"subscriber_id", newSubscriber.Data.ID,
		"email", newSubscriber.Data.Email,
	)

	return newSubscriber.Data.ID, nil
}

// AddSubscriberToGroup adds a subscriber to a specific group
func (s *Service) AddSubscriberToGroup(ctx context.Context, subscriberID, groupID string) error {
	if s.client == nil {
		return fmt.Errorf("MailerLite client not initialized")
	}

	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	s.log.Info(ctx, "Adding subscriber to MailerLite group",
		"subscriber_id", subscriberID,
		"group_id", groupID,
	)

	_, _, err := s.client.Group.Assign(ctx, groupID, subscriberID)
	if err != nil {
		s.log.Error(ctx, "Failed to assign subscriber to MailerLite group",
			"error", err,
			"subscriber_id", subscriberID,
			"group_id", groupID,
		)
		return fmt.Errorf("failed to assign subscriber to MailerLite group: %w", err)
	}

	s.log.Info(ctx, "Successfully assigned subscriber to MailerLite group",
		"subscriber_id", subscriberID,
		"group_id", groupID,
	)

	return nil
}
