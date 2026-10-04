# Notification HTTP contract

First-party inbox routes use the authenticated user and current workspace;
they are separate from the public developer `/api/v1` OpenAPI contract.

`PUT /workspaces/{workspaceSlug}/notifications/{id}/read` accepts an optional
single `observedCreatedAt` query parameter. Pass the displayed notification's
`createdAt` value unchanged, as a URL-encoded RFC3339 timestamp with its original
fractional precision. The server marks that version read only when its timestamp
matches the current row. Refreshing an unread story notification changes this
timestamp, so an older displayed row or push cannot consume the refreshed event.

- `204`: that version was marked read (including an already-read version).
- `400`: the parameter is malformed, empty, repeated, or exceeds 64 bytes.
- `409`: the authorized notification was refreshed after the supplied version;
  refresh the inbox and retain the new notification as unread.

Omitting the parameter preserves existing callers' behavior. Bulk mark-all-read
intentionally operates on all currently eligible rows and takes no version.
