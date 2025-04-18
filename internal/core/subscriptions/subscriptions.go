package subscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrInvoiceNotFound          = errors.New("invoice not found")
	ErrInvalidSubscription      = errors.New("invalid subscription data")
	ErrInvalidInvoice           = errors.New("invalid invoice data")
	ErrStripeOperationFailed    = errors.New("stripe operation failed")
	ErrAlreadyProcessingEvent   = errors.New("already processing event")
	ErrSubscriptionItemNotFound = errors.New("stripe subscription item ID not found for workspace")
)

// Repository defines methods for accessing subscription data
type Repository interface {
	GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error)
	CreateSubscription(ctx context.Context, subscription CoreWorkspaceSubscription) error
	UpdateSubscription(ctx context.Context, subscription CoreWorkspaceSubscription) error
	GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error)
	GetWorkspaceUserCount(ctx context.Context, workspaceID uuid.UUID) (int, error)

	// New methods for Stripe integration
	SaveStripeCustomerID(ctx context.Context, workspaceID uuid.UUID, customerID string) error
	UpdateSubscriptionDetails(ctx context.Context, workspaceID uuid.UUID, subID, itemID string, status SubscriptionStatus, seatCount int, trialEnd *time.Time, tier SubscriptionTier) error
	UpdateSubscriptionStatus(ctx context.Context, workspaceID uuid.UUID, subID string, status SubscriptionStatus) error
	UpdateSubscriptionSeatCount(ctx context.Context, workspaceID uuid.UUID, subID string, seatCount int) error
	CreateInvoice(ctx context.Context, invoice CoreSubscriptionInvoice) error
	HasEventBeenProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventAsProcessed(ctx context.Context, eventID string, eventType string, workspaceID *uuid.UUID, payload []byte) error
}

// Service provides subscription operations
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New creates a new subscription service
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// GetSubscription returns a subscription for a workspace
func (s *Service) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error) {
	return s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
}

// CreateSubscription creates a new subscription
func (s *Service) CreateSubscription(ctx context.Context, subscription CoreWorkspaceSubscription) error {
	return s.repo.CreateSubscription(ctx, subscription)
}

// UpdateSubscription updates an existing subscription
func (s *Service) UpdateSubscription(ctx context.Context, subscription CoreWorkspaceSubscription) error {
	return s.repo.UpdateSubscription(ctx, subscription)
}

// GetInvoices retrieves invoices for a workspace
func (s *Service) GetInvoices(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error) {
	return s.repo.GetInvoicesByWorkspaceID(ctx, workspaceID)
}
