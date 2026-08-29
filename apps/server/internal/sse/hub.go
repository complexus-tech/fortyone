package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	h.listeners.Add(3)
	go h.runClientListener(func() { h.listenToUserNotifications(client) })
	go h.runClientListener(func() { h.listenToUserUpdates(client) })
	go h.runClientListener(func() { h.listenToWorkspaceUpdates(client) })
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
	}
	select {
	case h.register <- client:
		return client, nil
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

// listenToUserNotifications is run in a goroutine for each connected client.
// It subscribes to the user-specific Redis Pub/Sub channel and forwards messages.
func (h *Hub) listenToUserNotifications(client *Client) {
	channelName := fmt.Sprintf("%s%s", pubSubChannelPrefix, client.UserID.String())
	pubsub := h.redisClient.Subscribe(client.ctx, channelName) // client.ctx will be cancelled on disconnect
	defer pubsub.Close()                                       // Ensure subscription is closed when this goroutine exits.

	// Wait for subscription to be confirmed.
	_, err := pubsub.Receive(client.ctx) // Use client.ctx for this operation as well
	if err != nil {
		// If client.ctx is already Done, this error might be expected (e.g., context canceled)
		if client.ctx.Err() != nil {
			h.log.Info(client.ctx, "Client disconnected before Pub/Sub subscription could be confirmed", "userID", client.UserID, "channel", channelName)
		} else {
			h.log.Error(client.ctx, "Failed to subscribe to Redis Pub/Sub channel", "userID", client.UserID, "channel", channelName, "error", err)
		}
		// Ensure client is unregistered if subscription fails and it's not already being unregistered
		// No need to call h.UnregisterClient(client) here because if subscription fails, the client goroutine exits,
		// and the HTTP handler is responsible for calling UnregisterClient upon detecting a disconnect or error.
		return
	}

	h.log.Info(client.ctx, "Subscribed to Redis Pub/Sub channel", "userID", client.UserID, "channel", channelName)
	redisChannel := pubsub.Channel() // Get the Go channel for messages

	for {
		select {
		case <-client.ctx.Done(): // Client disconnected (cancelled by unregister or app shutdown)
			h.log.Info(client.ctx, "Client context done, stopping Pub/Sub listener", "userID", client.UserID, "channel", channelName)
			return
		case msg, ok := <-redisChannel:
			if !ok { // Channel closed by Redis client (e.g. connection issue, or pubsub.Close() called)
				h.log.Info(client.ctx, "Redis Pub/Sub channel closed by library", "userID", client.UserID, "channel", channelName)
				// No need to call h.UnregisterClient here; the HTTP handler should detect the client has gone.
				return
			}

			// Attempt to deserialize the payload into a CoreNotification to make sure it's valid and for logging.
			var notificationPayload notifications.CoreNotification
			if err := json.Unmarshal([]byte(msg.Payload), &notificationPayload); err != nil {
				h.log.Error(client.ctx, "Failed to unmarshal notification from Pub/Sub, skipping", "userID", client.UserID, "channel", channelName, "payload", msg.Payload, "error", err)
				continue // Skip malformed messages
			}
			if !notificationMatchesWorkspace(notificationPayload, client.WorkspaceID) {
				h.log.Warn(client.ctx, "Dropping notification for a different workspace", "userID", client.UserID, "clientWorkspaceID", client.WorkspaceID, "notificationWorkspaceID", notificationPayload.WorkspaceID, "notificationID", notificationPayload.ID)
				continue
			}
			publicPayload, err := json.Marshal(notificationPayload.Public())
			if err != nil {
				h.log.Error(client.ctx, "Failed to sanitize notification from Pub/Sub, skipping", "userID", client.UserID, "channel", channelName, "notificationID", notificationPayload.ID, "error", err)
				continue
			}

			h.log.Debug(client.ctx, "Received message from Pub/Sub", "userID", client.UserID, "channel", channelName, "notificationID", notificationPayload.ID)

			// Send message to client's personal channel.
			select {
			case client.Send <- publicPayload:
				h.log.Debug(client.ctx, "Message sent to client's send channel", "userID", client.UserID, "notificationID", notificationPayload.ID)
			case <-time.After(clientSendTimeout):
				h.log.Warn(client.ctx, "Timeout sending message to client channel, client might be slow or send channel full", "userID", client.UserID, "channel", channelName, "notificationID", notificationPayload.ID)
			// Note: If client.Send is full and times out, the message is dropped for this client.
			// Consider strategies if this becomes an issue (e.g., increasing buffer, or more aggressive client disconnect).
			case <-client.ctx.Done(): // Check again in case client disconnected while trying to send.
				h.log.Info(client.ctx, "Client disconnected while attempting to send message from Pub/Sub", "userID", client.UserID, "channel", channelName)
				return
			}
		}
	}
}

// listenToUserUpdates forwards non-notification, user-scoped application
// updates such as a completed calendar sync to every active tab for that user.
func (h *Hub) listenToUserUpdates(client *Client) {
	channelName := publisher.UserUpdatesChannel(client.UserID)
	pubsub := h.redisClient.Subscribe(client.ctx, channelName)
	defer pubsub.Close()

	if _, err := pubsub.Receive(client.ctx); err != nil {
		if client.ctx.Err() == nil {
			h.log.Error(client.ctx, "Failed to subscribe to user updates", "userID", client.UserID, "channel", channelName, "error", err)
		}
		return
	}

	redisChannel := pubsub.Channel()
	for {
		select {
		case <-client.ctx.Done():
			return
		case msg, ok := <-redisChannel:
			if !ok {
				return
			}
			var update publisher.UserUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				h.log.Error(client.ctx, "Failed to unmarshal user update", "userID", client.UserID, "channel", channelName, "error", err)
				continue
			}
			if !userUpdateMatchesClient(update, client) {
				h.log.Warn(client.ctx, "Dropping user update outside the active SSE scope", "userID", client.UserID, "workspaceID", client.WorkspaceID, "updateUserID", update.UserID, "updateWorkspaceID", update.WorkspaceID)
				continue
			}
			select {
			case client.Send <- []byte(msg.Payload):
			case <-time.After(clientSendTimeout):
				h.log.Warn(client.ctx, "Timeout sending user update", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
			case <-client.ctx.Done():
				return
			}
		}
	}
}

// listenToWorkspaceUpdates is run in a goroutine for each connected client.
// It subscribes to the workspace-specific Redis Pub/Sub channel and forwards messages.
func (h *Hub) listenToWorkspaceUpdates(client *Client) {
	channelName := fmt.Sprintf("%s%s", workspaceChannelPrefix, client.WorkspaceID.String())
	pubsub := h.redisClient.Subscribe(client.ctx, channelName) // client.ctx will be cancelled on disconnect
	defer pubsub.Close()                                       // Ensure subscription is closed when this goroutine exits.

	// Wait for subscription to be confirmed.
	_, err := pubsub.Receive(client.ctx)
	if err != nil {
		// If client.ctx is already Done, this error might be expected (e.g., context canceled)
		if client.ctx.Err() != nil {
			h.log.Info(client.ctx, "Client disconnected before workspace Pub/Sub subscription could be confirmed", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
		} else {
			h.log.Error(client.ctx, "Failed to subscribe to workspace Redis Pub/Sub channel", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName, "error", err)
		}
		return
	}

	h.log.Info(client.ctx, "Subscribed to workspace Redis Pub/Sub channel", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
	redisChannel := pubsub.Channel() // Get the Go channel for messages

	for {
		select {
		case <-client.ctx.Done(): // Client disconnected (cancelled by unregister or app shutdown)
			h.log.Info(client.ctx, "Client context done, stopping workspace Pub/Sub listener", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
			return
		case msg, ok := <-redisChannel:
			if !ok { // Channel closed by Redis client (e.g. connection issue, or pubsub.Close() called)
				h.log.Info(client.ctx, "Workspace Redis Pub/Sub channel closed by library", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
				return
			}

			h.log.Debug(client.ctx, "Received workspace update from Pub/Sub", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)

			// Send workspace update to client's channel.
			select {
			case client.Send <- []byte(msg.Payload):
				h.log.Debug(client.ctx, "Workspace update sent to client's send channel", "userID", client.UserID, "workspaceID", client.WorkspaceID)
			case <-time.After(clientSendTimeout):
				h.log.Warn(client.ctx, "Timeout sending workspace update to client channel", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
			case <-client.ctx.Done(): // Check again in case client disconnected while trying to send.
				h.log.Info(client.ctx, "Client disconnected while attempting to send workspace update", "userID", client.UserID, "workspaceID", client.WorkspaceID, "channel", channelName)
				return
			}
		}
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
