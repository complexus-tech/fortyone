package subscriptions

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82/client"
)

var (
	ErrSubscriptionNotFound         = subscriptionsdomain.ErrSubscriptionNotFound
	ErrInvalidSubscription          = errors.New("invalid subscription data")
	ErrInvalidInvoice               = errors.New("invalid invoice data")
	ErrStripeOperationFailed        = errors.New("stripe operation failed")
	ErrAlreadyProcessingEvent       = errors.New("already processing event")
	ErrSubscriptionItemNotFound     = errors.New("stripe subscription item ID not found for workspace")
	ErrWorkspaceHasActiveSub        = errors.New("workspace already has an active subscription, use change plan flow")
	ErrAlreadySubscribedToThisPlan  = errors.New("already subscribed to this specific plan")
	ErrNoActiveSubscriptionToChange = errors.New("no active subscription found to change")
	ErrSubscriptionAlreadyCanceled  = errors.New("subscription is already canceled or pending cancellation")
	ErrInvalidBillingRedirect       = errors.New("billing redirect URL is not allowed")
	ErrInvalidPriceLookupKey        = errors.New("price lookup key is not supported")
)

// Repository is the subscriptions application's persistence boundary.
// Provider-facing writes include both the tenant identity and the immutable
// Stripe identity so the repository can enforce their binding atomically.
type Repository interface {
	WebhookRepository

	GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error)
	HasActiveSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (bool, error)
	GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error)
	GetWorkspaceUserCount(ctx context.Context, workspaceID uuid.UUID) (int, error)
	GetWorkspaceCreatorEmail(ctx context.Context, workspaceID uuid.UUID) (string, error)
	UpdateWorkspaceSubscription(ctx context.Context, workspaceID uuid.UUID, snapshot subscriptionsdomain.SubscriptionSnapshot) error
	ApplyStripeSubscriptionSnapshot(ctx context.Context, snapshot subscriptionsdomain.SubscriptionSnapshot, cursor subscriptionsdomain.StripeEventCursor) (subscriptionsdomain.SubscriptionMutation, error)
	UpsertStripeSubscription(ctx context.Context, workspaceID uuid.UUID, snapshot subscriptionsdomain.SubscriptionSnapshot, cursor subscriptionsdomain.StripeEventCursor) (subscriptionsdomain.SubscriptionMutation, error)
	ApplyStripeSubscriptionDeletion(ctx context.Context, subscriptionID string, cursor subscriptionsdomain.StripeEventCursor) (subscriptionsdomain.SubscriptionMutation, error)
	UpsertStripeInvoice(ctx context.Context, customerID string, invoice CoreSubscriptionInvoice) error
}

// Service coordinates local subscription state with Stripe.
type Service struct {
	repo           Repository
	webhookRepo    WebhookRepository
	webhookEvents  webhookEventProcessor
	log            *logger.Logger
	stripeClient   *client.API
	tasksService   *tasks.Service
	webhookSecret  string
	webhookClock   func() time.Time
	webhookLease   time.Duration
	redirectOrigin *url.URL
}

type Option func(*Service) error

func WithRedirectOrigin(rawURL string) Option {
	return func(service *Service) error {
		if strings.TrimSpace(rawURL) == "" {
			return nil
		}
		origin, err := parseBillingOrigin(rawURL)
		if err != nil {
			return err
		}
		service.redirectOrigin = origin
		return nil
	}
}

// New constructs a subscription service. Stripe is required because every
// state-changing operation reconciles against provider state.
func New(log *logger.Logger, repo Repository, stripeClient *client.API, webhookSecret string, tasksService *tasks.Service, options ...Option) *Service {
	if stripeClient == nil {
		panic("Stripe client cannot be nil")
	}

	service := &Service{
		repo:          repo,
		webhookRepo:   repo,
		log:           log,
		stripeClient:  stripeClient,
		webhookSecret: webhookSecret,
		tasksService:  tasksService,
		webhookClock:  time.Now,
		webhookLease:  defaultWebhookEventLease,
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(service); err != nil {
			panic("invalid subscriptions service option: " + err.Error())
		}
	}

	service.webhookEvents = serviceWebhookEventProcessor{service: service}
	return service
}
