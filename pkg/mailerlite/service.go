package mailerlite

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/mailerlite/mailerlite-go"
)

// Service wraps the MailerLite client with additional functionality
type Service struct {
	client            *mailerlite.Client
	log               *logger.Logger
	onboardingGroupID string
}

// Config holds the configuration for MailerLite service
type Config struct {
	APIKey            string
	OnboardingGroupID string
}

// NewService creates a new MailerLite service instance
func NewService(log *logger.Logger, config Config) *Service {
	if config.APIKey == "" {
		return nil
	}
	client := mailerlite.NewClient(config.APIKey)
	return &Service{
		client:            client,
		log:               log,
		onboardingGroupID: config.OnboardingGroupID,
	}
}

// AddToOnboardingGroup creates/updates a subscriber and adds them to the onboarding group
func (s *Service) AddToOnboardingGroup(ctx context.Context, email, fullName string) error {
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
		return fmt.Errorf("failed to create/update MailerLite subscriber: %w", err)
	}

	s.log.Info(ctx, "MailerLite subscriber created/updated",
		"subscriber_id", newSubscriber.Data.ID,
		"email", newSubscriber.Data.Email,
	)

	// Add subscriber to onboarding group
	s.log.Info(ctx, "Adding subscriber to MailerLite onboarding group",
		"subscriber_id", newSubscriber.Data.ID,
		"group_id", s.onboardingGroupID,
	)

	_, _, err = s.client.Group.Assign(ctx, s.onboardingGroupID, newSubscriber.Data.ID)
	if err != nil {
		s.log.Error(ctx, "Failed to assign subscriber to MailerLite group",
			"error", err,
			"subscriber_id", newSubscriber.Data.ID,
			"group_id", s.onboardingGroupID,
		)
		return fmt.Errorf("failed to assign subscriber to MailerLite group: %w", err)
	}

	s.log.Info(ctx, "Successfully assigned subscriber to MailerLite onboarding group",
		"subscriber_id", newSubscriber.Data.ID,
		"group_id", s.onboardingGroupID,
	)

	return nil
}

// IsConfigured returns true if the MailerLite service is properly configured
func (s *Service) IsConfigured() bool {
	return s != nil && s.client != nil && s.onboardingGroupID != ""
}
