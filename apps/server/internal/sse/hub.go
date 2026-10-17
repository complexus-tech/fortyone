package sse

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// pubSubChannelPrefix is the prefix for user-specific notification channels.
	pubSubChannelPrefix = "user-notifications:"
	// workspaceChannelPrefix is the prefix for workspace-wide update channels.
	workspaceChannelPrefix = "workspace-updates:"
	// clientSendTimeout is the timeout for sending a message to a client's channel.
	clientSendTimeout = 1 * time.Second
)

// Client represents an active SSE client connection.
type Client struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID   // Track workspace for workspace-wide updates
	Send        chan []byte // Channel for sending messages to this client.
	ctx         context.Context
	cancelFunc  context.CancelFunc
	ready       chan error
}

// Ctx returns the client's context. This context is cancelled when the client unregisters or the hub shuts down.
func (c *Client) Ctx() context.Context {
	return c.ctx
}

// Hub manages all active SSE clients and their Redis Pub/Sub subscriptions.
type Hub struct {
	redisClient *redis.Client
	log         *logger.Logger

	lifecycleMu sync.RWMutex
	appCtx      context.Context
	running     bool
	listeners   sync.WaitGroup

	// mu protects the clients map.
	mu sync.RWMutex
	// clients is a map of user IDs to a map of their active clients.
	// A user might have multiple SSE connections (e.g., multiple browser tabs).
	clients map[uuid.UUID]map[*Client]bool

	register   chan *Client
	unregister chan *Client
}

func notificationMatchesWorkspace(notification notifications.CoreNotification, workspaceID uuid.UUID) bool {
	return workspaceID != uuid.Nil && notification.WorkspaceID == workspaceID
}

func userUpdateMatchesClient(update publisher.UserUpdate, client *Client) bool {
	return client != nil && update.UserID == client.UserID && update.WorkspaceID == client.WorkspaceID
}

// NewHub creates a new Hub instance. Run must be called by the process
// supervisor before clients can register.
func NewHub(log *logger.Logger, redisClient *redis.Client) *Hub {
	return &Hub{
		redisClient: redisClient,
		log:         log,
		clients:     make(map[uuid.UUID]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

// Run owns the Hub's main loop and every per-client Redis listener. It returns
// only after the supplied context is cancelled and all listeners have exited.
func (h *Hub) Run(ctx context.Context) error {
	if err := h.Initialize(ctx); err != nil {
		return err
	}

	h.lifecycleMu.Lock()
	if h.running {
		h.lifecycleMu.Unlock()
		return errors.New("SSE hub is already running")
	}
	h.appCtx = ctx
	h.running = true
	h.lifecycleMu.Unlock()

	h.log.Info(ctx, "SSE Hub starting")
	defer func() {
		h.lifecycleMu.Lock()
		h.running = false
		h.lifecycleMu.Unlock()

		h.shutdownAllClients(ctx)
		h.listeners.Wait()

		h.lifecycleMu.Lock()
		h.appCtx = nil
		h.lifecycleMu.Unlock()
		h.log.Info(context.Background(), "SSE Hub stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case client := <-h.register:
			h.handleRegistration(client)
		case client := <-h.unregister:
			h.handleUnregistration(client)
		}
	}
}

// Initialize validates that the hub can participate in a supervised run. Redis
// reachability is checked by process readiness and by startup before this hub is
// constructed; Run owns only subscription lifecycle.
func (h *Hub) Initialize(ctx context.Context) error {
	if h == nil || h.redisClient == nil {
		return errors.New("SSE hub requires a Redis client")
	}
	if h.log == nil {
		return errors.New("SSE hub requires a logger")
	}
	if ctx == nil {
		return errors.New("SSE hub context is required")
	}
	h.lifecycleMu.RLock()
	defer h.lifecycleMu.RUnlock()
	if h.running {
		return errors.New("SSE hub is already running")
	}
	return nil
}

func (h *Hub) handleRegistration(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true
	h.mu.Unlock()

	h.log.Info(client.ctx, "SSE client registered", "userID", client.UserID, "workspaceID", client.WorkspaceID)

	// All listeners are accounted for so Run cannot return while a Redis
	// subscription still owns process resources.
	h.listeners.Add(1)
	go h.runClientListener(func() { h.listenToClientEvents(client) })
}

func (h *Hub) handleUnregistration(client *Client) {
	client.cancelFunc()

	h.mu.Lock()
	if userClients, ok := h.clients[client.UserID]; ok {
		if _, clientExists := userClients[client]; clientExists {
			delete(userClients, client)
			if len(userClients) == 0 {
				delete(h.clients, client.UserID)
			}
			h.log.Info(client.ctx, "SSE client unregistered", "userID", client.UserID)
		}
	}
	h.mu.Unlock()
}

// RegisterNewClient is called by the SSE HTTP handler to register a new client.
// It creates a new client context that can be cancelled when the client disconnects.
func (h *Hub) RegisterNewClient(userID, workspaceID uuid.UUID) (*Client, error) {
	appCtx, running := h.lifecycleContext()
	if !running || appCtx == nil {
		return nil, errors.New("SSE hub is not accepting clients")
	}

	clientCtx, cancel := context.WithCancel(appCtx)
	client := &Client{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Send:        make(chan []byte, 256), // Buffered channel.
		ctx:         clientCtx,
		cancelFunc:  cancel,
		ready:       make(chan error, 1),
	}
	select {
	case h.register <- client:
		// Do not acknowledge an SSE connection until every Redis channel is ready.
		// Otherwise the browser's reconnect refetch can race subscription setup.
		select {
		case err := <-client.ready:
			if err != nil {
				h.UnregisterClient(client)
				return nil, err
			}
			return client, nil
		case <-clientCtx.Done():
			h.UnregisterClient(client)
			return nil, errors.New("SSE subscription was cancelled")
		}
	case <-appCtx.Done():
		cancel()
		return nil, errors.New("SSE hub is shutting down")
	}
}

// UnregisterClient is called by the SSE HTTP handler when a client disconnects.
func (h *Hub) UnregisterClient(client *Client) {
	if client == nil {
		return
	}
	client.cancelFunc()
	appCtx, running := h.lifecycleContext()
	if !running || appCtx == nil {
		return
	}
	select {
	case h.unregister <- client:
	case <-appCtx.Done():
	}
}

func (h *Hub) lifecycleContext() (context.Context, bool) {
	h.lifecycleMu.RLock()
	defer h.lifecycleMu.RUnlock()
	return h.appCtx, h.running
}

func (h *Hub) runClientListener(listener func()) {
	defer h.listeners.Done()
	listener()
}

// A single Redis socket carries the three existing scopes for a browser.
// Keeping subscriptions per client preserves independent cancellation and tab
// lifetimes while avoiding three sockets and listener trees per connection.
func (h *Hub) listenToClientEvents(client *Client) {
	defer client.cancelFunc()
	notificationChannel := pubSubChannelPrefix + client.UserID.String()
	updateChannel := publisher.UserUpdatesChannel(client.UserID)
	workspaceChannel := workspaceChannelPrefix + client.WorkspaceID.String()
	channels := []string{notificationChannel, updateChannel, workspaceChannel}
	pubsub := h.redisClient.Subscribe(client.ctx, channels...)
	defer pubsub.Close()
	deadline := time.Now().Add(5 * time.Second)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			client.ready <- errors.New("SSE subscription setup timed out")
			return
		}
		reply, err := pubsub.ReceiveTimeout(client.ctx, remaining)
		if err != nil {
			if client.ctx.Err() == nil {
				h.log.Error(client.ctx, "Failed to subscribe SSE client", "userID", client.UserID, "workspaceID", client.WorkspaceID, "error", err)
			}
			client.ready <- errors.New("SSE subscription is unavailable")
			return
		}
		if subscription, ok := reply.(*redis.Subscription); ok && subscription.Kind == "subscribe" && subscription.Count == len(channels) {
			break
		}
	}
	client.ready <- nil
	// Redis Pub/Sub is not a durable log. A reconnect can miss hints while the
	// browser still believes its SSE connection is healthy. Surface resubscribe
	// acknowledgements so we can close that browser stream and reconcile from
	// authoritative HTTP state after EventSource reconnects.
	messages := pubsub.ChannelWithSubscriptions()
	for {
		select {
		case <-client.ctx.Done():
			return
		case event, ok := <-messages:
			if !ok {
				return
			}
			if _, reconnected := event.(*redis.Subscription); reconnected {
				h.log.Info(client.ctx, "Redis resubscribed; closing SSE client for reconciliation", "userID", client.UserID, "workspaceID", client.WorkspaceID)
				return
			}
			message, ok := event.(*redis.Message)
			if !ok {
				continue
			}
			payload, valid := h.clientPayload(client, message, notificationChannel, updateChannel, workspaceChannel)
			if !valid {
				continue
			}
			timer := time.NewTimer(clientSendTimeout)
			select {
			case client.Send <- payload:
				timer.Stop()
			case <-timer.C:
				// Closing forces the normal reconnect/history reconciliation path. A
				// connected stream that silently drops hints can leave stale UI forever.
				h.log.Warn(client.ctx, "Closing slow SSE client", "userID", client.UserID, "workspaceID", client.WorkspaceID)
				return
			case <-client.ctx.Done():
				timer.Stop()
				return
			}
		}
	}
}

func (h *Hub) clientPayload(client *Client, message *redis.Message, notificationChannel, updateChannel, workspaceChannel string) ([]byte, bool) {
	switch message.Channel {
	case notificationChannel:
		var notification notifications.CoreNotification
		if err := json.Unmarshal([]byte(message.Payload), &notification); err != nil || notification.RecipientID != client.UserID || !notificationMatchesWorkspace(notification, client.WorkspaceID) {
			return nil, false
		}
		payload, err := json.Marshal(notification.Public())
		if err != nil {
			return nil, false
		}
		return payload, true
	case updateChannel:
		var update publisher.UserUpdate
		if err := json.Unmarshal([]byte(message.Payload), &update); err != nil || !userUpdateMatchesClient(update, client) {
			return nil, false
		}
		return []byte(message.Payload), true
	case workspaceChannel:
		return []byte(message.Payload), true
	default:
		return nil, false
	}
}

// shutdownAllClients is called when the application is shutting down.
func (h *Hub) shutdownAllClients(ctx context.Context) {
	h.log.Info(ctx, "Shutting down all SSE clients")
	h.mu.Lock() // Lock to safely iterate and modify
	defer h.mu.Unlock()

	for userID, userClients := range h.clients {
		for client := range userClients {
			client.cancelFunc()
			h.log.Debug(ctx, "Signaled client to shutdown", "userID", userID)
		}
		delete(h.clients, userID) // Remove the user entry from the clients map
	}
	h.clients = make(map[uuid.UUID]map[*Client]bool) // Re-initialize to clear map completely
	h.log.Info(ctx, "All SSE clients signaled for shutdown and hub cleared")
}

// BroadcastToUser (Optional utility, if direct broadcast from hub is ever needed, though primary path is via Redis Pub/Sub)
// This is NOT the primary way notifications will be sent. Primary path is Consumer -> Redis Pub/Sub -> Hub's listenToPubSub.
// This is more for system messages or if the hub itself needed to send something.
func (h *Hub) BroadcastToUser(userID uuid.UUID, message []byte) {
	h.mu.RLock()
	userClients, ok := h.clients[userID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	// Create a new slice for userClients to avoid holding lock while sending
	// This is a common pattern to avoid holding a lock during potentially blocking send operations.
	clientsToSend := make([]*Client, 0, len(userClients))
	for client := range userClients {
		clientsToSend = append(clientsToSend, client)
	}
	h.mu.RUnlock()

	for _, client := range clientsToSend {
		select {
		case client.Send <- message:
			h.log.Debug(client.ctx, "Direct broadcast message sent to client", "userID", userID)
		case <-time.After(clientSendTimeout):
			h.log.Warn(client.ctx, "Timeout sending direct broadcast message to client", "userID", userID)
		case <-client.ctx.Done():
			// Client context is done, probably disconnected
			h.log.Info(client.ctx, "Client disconnected during direct broadcast attempt", "userID", userID)
		}
	}
}
