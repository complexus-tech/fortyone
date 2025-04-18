package subscriptions

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvalidSubscription  = errors.New("invalid subscription data")
	ErrInvalidInvoice       = errors.New("invalid invoice data")
)

// Repository defines methods for accessing subscription data
type Repository interface {
	GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error)
	CreateSubscription(ctx context.Context, subscription CoreWorkspaceSubscription) error
	UpdateSubscription(ctx context.Context, subscription CoreWorkspaceSubscription) error
	GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error)
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
