# Real-time Notifications with Server-Sent Events (SSE) in Go

This document explains how real-time notifications are implemented in this Go project using Server-Sent Events (SSE) and Redis Pub/Sub.

## 1. What are Real-time Notifications?

Real-time notifications allow the server to send updates to users (clients) as soon as events happen, without the user needing to refresh their page or manually ask for new data. For example, when one user assigns a story to another, the assigned user should receive an immediate notification in the web application.

## 2. Chosen Technology: Server-Sent Events (SSE)

### What is SSE?

Server-Sent Events (SSE) is a technology that enables a web server to push data to a client (typically a web browser) once an initial client connection has been established. It's a standard HTML5 API. The key characteristic is that communication is **one-way**: from server to client.

### Why SSE for this project?

- **Simplicity**: For sending notifications _from_ the server _to_ the client, SSE is often simpler to implement than WebSockets, which are designed for full bi-directional communication.
- **Automatic Reconnection**: Browsers that support SSE will automatically attempt to reconnect if the connection is lost, which is great for reliability.
- **HTTP-based**: SSE works over standard HTTP, making it compatible with most existing infrastructure.

The client-side uses the `EventSource` JavaScript API to connect to an SSE endpoint.

## 3. The Role of Redis Pub/Sub

### What is Redis Pub/Sub?

Redis Pub/Sub (Publish/Subscribe) is a messaging paradigm where senders (publishers) send messages to channels, and receivers (subscribers) subscribe to channels they are interested in, receiving messages sent to those channels without direct knowledge of the publishers.

### Why do we need it with SSE?

1.  **Decoupling**: The part of our system that creates a notification (e.g., when a story is updated) doesn't need to know which specific users are currently connected via SSE or how to send them the message directly. It just publishes the notification event to a Redis channel.
2.  **Scalability**: If our Go application runs in multiple instances (e.g., multiple Docker containers), any instance can publish a notification. A central SSE Hub (explained below) can then subscribe to Redis and forward messages to the correct clients, regardless of which server instance the client is connected to or which instance published the event.
3.  **Targeted Notifications**: We can use user-specific channel names (e.g., `user-notifications:USER_ID_XYZ`) so that the SSE Hub only processes and forwards messages relevant to its connected clients.

## 4. Backend Implementation - The Big Picture

Here's a simplified flow of how a notification gets from an event in the system to the user:

1.  **Event Occurs**: E.g., User A assigns a story to User B.
2.  **Notification Created**: The system creates a notification record in the database for User B.
3.  **Publish to Redis**: After successful DB creation, the system (specifically, the `pkg/consumer/consumer.go` after processing a stream event) publishes the notification data (as JSON) to a specific Redis channel, e.g., `user-notifications:USER_B_ID`.
4.  **SSE Hub Receives**: Our `SSE Hub` (running in the Go backend) is subscribed to Redis channels for all currently connected users. If User B is connected, their dedicated subscription in the Hub picks up the message from `user-notifications:USER_B_ID`.
5.  **Forward to Client**: The `SSE Hub` sends this message to the specific HTTP connection handler for User B.
6.  **Client Receives**: The HTTP handler writes the message in the SSE format over the persistent HTTP connection. The client-side `EventSource` receives the data and can display the notification.

## 5. Core Go Components

### a. The SSE Hub (`internal/sse/hub.go`)

This is the heart of the SSE server-side logic.

**Responsibilities**:

- Manages all active SSE client connections.
- Keeps track of which user is associated with which connection.
- Handles client registrations (when a user connects) and unregistrations (when they disconnect).
- For each registered client (or more accurately, for each user with active clients), it maintains a subscription to their dedicated Redis Pub/Sub channel (e.g., `user-notifications:<userID>`).
- Listens for messages on these Redis channels.
- When a message arrives from Redis, it forwards the message to the appropriate client(s) connected for that user.
- Manages graceful shutdown of client connections.

**Key Parts**:

- `Client` struct: Represents an active SSE client, holding their `UserID`, a Go channel (`Send`) to push messages to them, and a `context.Context` for managing their lifecycle.
- `Hub` struct: Holds references to the Redis client, logger, active clients (typically a `map[uuid.UUID]map[*Client]bool` to support multiple connections per user), and channels for registration/unregistration requests.
- `Run()`: The main loop of the Hub, processing registrations, unregistrations, and application shutdown signals.
- `handleRegistration()` / `handleUnregistration()`: Logic for adding/removing clients.
- `RegisterNewClient()` / `UnregisterClient()`: Methods called by the HTTP handler to interact with the Hub.
- `listenToPubSub()`: A goroutine started for each user (or client, depending on design) that subscribes to their Redis channel and forwards messages to the client's `Send` channel.

### b. The SSE HTTP Handler (`internal/handlers/ssegrp/handler.go`)

This is the Go HTTP handler that clients connect to establish an SSE connection.

**Responsibilities**:

- Handles incoming HTTP GET requests to the SSE endpoint (e.g., `/api/v1/notifications/subscribe`).
- **Authentication**: Uses existing authentication middleware (`mid.Auth`) to identify the user making the request.
- **Set SSE Headers**: Sets necessary HTTP headers for an SSE stream:
  - `Content-Type: text/event-stream`
  - `Cache-Control: no-cache`
  - `Connection: keep-alive`
  - `Access-Control-Allow-Origin`: For Cross-Origin Resource Sharing (CORS), crucial if your frontend is on a different domain.
- **Client Registration**: Registers the authenticated user (and their new connection) with the `SSE Hub`.
- **Keep-Alive**: Periodically sends SSE comments (e.g., `:keep-alive

`) to prevent the connection from timing out.

- **Event Loop**:
  - Listens for messages on the client's `Send` channel (populated by the `Hub` from Redis).
  - Formats these messages according to the SSE specification (e.g., `data: <json_payload>

`) and writes them to the `http.ResponseWriter`.
    -   Listens for client disconnect signals (e.g., `r.Context().Done()`if the browser closes the connection, or`sseClient.Ctx().Done()` if the Hub cancels the client).

- **Client Unregistration**: Ensures the client is unregistered from the `Hub` when the connection closes for any reason.

**Key Parts**:

- `Handler` struct: Holds dependencies like the logger and the `SSEHub`.
- `StreamNotifications()`: The method that implements `http.Handler` (or `web.Handler` in this project's framework).
  - Retrieves `userID` from context (set by auth middleware).
  - Sets HTTP headers.
  - Obtains an `http.Flusher` to push data incrementally.
  - Calls `h.SSEHub.RegisterNewClient(userID)`.
  - Sends an initial "connected" event (optional).
  - Enters a `for/select` loop to:
    - Listen on `r.Context().Done()` (client closed request).
    - Listen on `sseClient.Ctx().Done()` (hub closed client context).
    - Listen on `sseClient.Send` for messages from the Hub.
    - Send keep-alive pings.
  - Calls `h.SSEHub.UnregisterClient(sseClient)` in a `defer` block.

### c. Redis Publishing (Conceptual - in `pkg/consumer/consumer.go`)

While the SSE components handle _delivering_ notifications, another part of the system is responsible for _generating_ them and putting them onto Redis. In this project, this is typically handled by the `pkg/consumer/consumer.go`.

When the consumer processes an event (e.g., from a Redis Stream indicating a story was updated) that should trigger a notification:

1.  It creates the notification in the database (via `notifications.Create(...)`).
2.  If successful and a `RecipientID` is present, it marshals the `CoreNotification` object into JSON.
3.  It constructs a Redis channel name: `fmt.Sprintf("user-notifications:%s", notification.RecipientID.String())`.
4.  It uses the Redis client to `PUBLISH` the JSON payload to this channel: `c.redis.Publish(ctx, channelName, notificationJSON)`.

## 6. Wiring It All Together (Configuration & Routing)

The SSE functionality needs to be integrated into the application's main setup:

- **`internal/handlers/ssegrp/routes.go`**:

  - Defines a `Config` struct for `ssegrp` dependencies (Logger, SecretKey for auth, SSEHub, CorsOrigin).
  - Has a `Routes` function that initializes the `ssegrp.Handler` and registers the `/api/v1/notifications/subscribe` GET route with the main web application, applying authentication middleware.

- **`internal/mux/mux.go`**:

  - The main `mux.Config` struct is extended to include `SSEHub *sse.Hub` and `CorsOrigin string` so these can be passed down from `main.go`.

- **`internal/handlers/handlers.go`**:

  - The `BuildAllRoutes` function (which sets up all application routes) is modified:
    - It takes the `mux.Config` (which now includes the `SSEHub` and `CorsOrigin`).
    - It initializes the `ssegrp.Config` using values from the main `mux.Config`.
    - It calls `ssegrp.Routes(...)` to set up the SSE-specific routes.

- **`cmd/api/main.go`**:
  - In the `run()` function:
    - An instance of `sse.Hub` is created: `sseHub := sse.NewHub(ctx, log, rdb)`. `ctx` here is the main application context for graceful shutdown.
    - The Hub's main processing loop is started in a new goroutine: `go sseHub.Run()`.
    - The `sseHub` instance and the desired `CorsOrigin` (e.g., `"*"` for development, or a specific frontend URL like `cfg.Web.CorsOrigin` for production) are passed into the `mux.Config`.

## 7. End-to-End Flow (Revisited)

1.  **App Start**: `main.go` initializes Redis, logger, DB, etc. It creates an `sse.Hub` instance and starts `sseHub.Run()` in a goroutine. The Hub is now ready.
2.  **Client Connects**: User opens the web app. Frontend JavaScript executes `new EventSource("/api/v1/notifications/subscribe")`.
3.  **HTTP Request**: Browser sends a GET request to `/api/v1/notifications/subscribe`.
4.  **Auth & Handler**:
    - Auth middleware (`mid.Auth`) in `ssegrp/routes.go` verifies the user and gets their ID.
    - `ssegrp.Handler.StreamNotifications` is called.
    - It sets SSE headers (Content-Type, Cache-Control, Connection, CORS).
    - It calls `hub.RegisterNewClient(userID)`.
5.  **Hub Registers Client**:
    - `hub.handleRegistration()` adds the new client to its internal map.
    - `hub.listenToPubSub()` starts a new goroutine for this client (or ensures one is running for this user), subscribing to `user-notifications:<userID>` on Redis.
6.  **Connection Open**: The `StreamNotifications` handler is now in its event loop, keeping the HTTP connection open.
7.  **Event Triggered**: An action in the app (e.g., story assignment) leads to the `consumer` creating a DB notification and then publishing it to `user-notifications:<userID>` on Redis.
8.  **Redis to Hub**: The `hub`'s `listenToPubSub` goroutine for that user receives the message from Redis.
9.  **Hub to Handler**: It sends the message payload to `client.Send` channel.
10. **Handler to Client**: The `StreamNotifications` handler's loop receives from `client.Send`, formats it as `data: ...

`, and writes it to the HTTP response.
11. **Client Receives**: The browser's `EventSource`fires an event (e.g.,`onmessage`or a custom event if`event:` field was used), and the frontend JS can now display the notification.

## 8. Client-Side (Briefly)

On the frontend, you would use the `EventSource` API:

```javascript
const userID = "USER_ID_FROM_AUTH_SOMEHOW"; // Or handled by cookie/auth header
const eventSource = new EventSource("/api/v1/notifications/subscribe"); // Assuming auth is cookie-based

eventSource.onopen = () => {
  console.log("SSE connection established.");
};

eventSource.onmessage = (event) => {
  console.log("Raw data:", event.data);
  try {
    const notification = JSON.parse(event.data);
    console.log("Received notification:", notification);
    // Process and display the notification
    // e.g., show a toast, update a badge
  } catch (e) {
    console.error("Failed to parse notification data:", e);
  }
};

// Example for custom named events (if server sends 'event: customEventName')
// eventSource.addEventListener('customEventName', (event) => {
//   console.log("Custom event data:", event.data);
// });

eventSource.onerror = (error) => {
  console.error("EventSource failed:", error);
  // The browser will automatically try to reconnect by default
  // You might want to close it explicitly under certain error conditions
  // if (eventSource.readyState === EventSource.CLOSED) {
  //   console.log("SSE connection was closed.");
  // } else if (eventSource.readyState === EventSource.CONNECTING) {
  //   console.log("SSE reconnecting...");
  // }
};

// To close the connection:
// eventSource.close();
```

This provides a comprehensive overview of the SSE notification system implemented.
