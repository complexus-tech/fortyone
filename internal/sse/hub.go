package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/pkg/logger"
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
	appCtx      context.Context // Main application context for graceful shutdown.

	// mu protects the clients map.
	mu sync.RWMutex
	// clients is a map of user IDs to a map of their active clients.
	// A user might have multiple SSE connections (e.g., multiple browser tabs).
	clients map[uuid.UUID]map[*Client]bool

	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new Hub instance.
func NewHub(ctx context.Context, log *logger.Logger, redisClient *redis.Client) *Hub {
	return &Hub{
		redisClient: redisClient,
		log:         log,
		appCtx:      ctx,
		clients:     make(map[uuid.UUID]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

// Run starts the Hub's main loop for managing client registrations and unregistrations.
// It should be run in a separate goroutine.
func (h *Hub) Run() {
	h.log.Info(h.appCtx, "SSE Hub starting...")
	defer h.log.Info(h.appCtx, "SSE Hub stopped.")

	for {
		select {
		case <-h.appCtx.Done(): // Application is shutting down
			h.shutdownAllClients()
			return
		case client := <-h.register:
			h.handleRegistration(client)
		case client := <-h.unregister:
			h.handleUnregistration(client)
		}
	}
}

func (h *Hub) handleRegistration(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true
	h.mu.Unlock()

	h.log.Info(client.ctx, "SSE client registered", "userID", client.UserID, "workspaceID", client.WorkspaceID)

	// Start listening to both user notifications and workspace updates
	go h.listenToUserNotifications(client)
	go h.listenToWorkspaceUpdates(client)
}

func (h *Hub) handleUnregistration(client *Client) {
	client.cancelFunc() // Signal the client's listenToPubSub goroutine to stop.

	h.mu.Lock()
	if userClients, ok := h.clients[client.UserID]; ok {
		if _, clientExists := userClients[client]; clientExists {
			delete(userClients, client)
			// It's important to close the Send channel only once.
			// Check if channel is already closed before attempting to close it if necessary,
			// though in this flow, unregister should only be called once per client.
			close(client.Send) // Close the send channel.
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
func (h *Hub) RegisterNewClient(userID, workspaceID uuid.UUID) *Client {
	clientCtx, cancel := context.WithCancel(h.appCtx) // Client context derived from app context.
	client := &Client{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Send:        make(chan []byte, 256), // Buffered channel.
		ctx:         clientCtx,
		cancelFunc:  cancel,
	}
	h.register <- client
	return client
}

// UnregisterClient is called by the SSE HTTP handler when a client disconnects.
func (h *Hub) UnregisterClient(client *Client) {
	if client == nil {
		return
	}
	h.unregister <- client
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

			h.log.Debug(client.ctx, "Received message from Pub/Sub", "userID", client.UserID, "channel", channelName, "notificationID", notificationPayload.ID)

			// Send message to client's personal channel.
			select {
			case client.Send <- []byte(msg.Payload):
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
func (h *Hub) shutdownAllClients() {
	h.log.Info(h.appCtx, "Shutting down all SSE clients...")
	h.mu.Lock() // Lock to safely iterate and modify
	defer h.mu.Unlock()

	for userID, userClients := range h.clients {
		for client := range userClients {
			client.cancelFunc() // Signal client's goroutine to stop
			close(client.Send)  // Close its send channel
			h.log.Debug(h.appCtx, "Signaled client to shutdown", "userID", userID)
		}
		delete(h.clients, userID) // Remove the user entry from the clients map
	}
	h.clients = make(map[uuid.UUID]map[*Client]bool) // Re-initialize to clear map completely
	h.log.Info(h.appCtx, "All SSE clients signaled for shutdown and hub cleared.")
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
