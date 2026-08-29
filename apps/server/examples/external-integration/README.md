# External integration sample

This runnable consumer imports only the public Go SDK preview. It demonstrates:

- PAT authentication and opaque-cursor story reads;
- an optional least-privilege service-account key for the idempotent story write;
- exact-byte outbound webhook verification before JSON decoding;
- a bounded request body and safe logs containing identifiers only;
- a file-backed durable demo inbox with `Webhook-Id` deduplication.

The local JSON-lines inbox makes the example restartable. A production
integration should replace it with a database transaction that inserts the
unique delivery ID and raw payload before returning `2xx`, then processes that
inbox asynchronously.

## Run it

From this directory:

```sh
export FORTYONE_TOKEN='<show-once PAT>'
export FORTYONE_WORKSPACE_ID='<workspace UUID>'
export FORTYONE_WEBHOOK_SECRET='<show-once whsec_ value>'
# Optional write demonstration. Set all three values together. Keep the
# idempotency key stable across retries and process restarts.
export FORTYONE_WRITE_TOKEN='<show-once service-account key>'
export FORTYONE_CREATE_STORY_TEAM_ID='<team UUID>'
export FORTYONE_CREATE_STORY_TITLE='Created by the external integration sample'
export FORTYONE_CREATE_STORY_IDEMPOTENCY_KEY='<at least 16 visible ASCII characters>'
go run .
```

Generate the optional value with `fortyone.NewIdempotencyKey` in a real Go
integration, then durably store it with the pending operation. This sample
accepts the key through the environment so restarting the process demonstrates
reuse instead of silently generating a new operation.

For the least-privilege path, create a service account in the workspace's
Developer settings, issue a key with only `stories:write`, restrict it to the
target team, and set that show-once value as `FORTYONE_WRITE_TOKEN`. The PAT in
`FORTYONE_TOKEN` needs only the read scopes used by the initial synchronization.
If `FORTYONE_WRITE_TOKEN` is absent, the sample keeps the backwards-compatible
PAT write path.

The receiver listens on `:8080` at `POST /webhooks/fortyone`. Set
`FORTYONE_LISTEN_ADDR` or `FORTYONE_DELIVERY_LOG` to change the local listener
or inbox. `FORTYONE_API_URL` is an optional HTTPS-compatible API proxy URL; it
must never contain credentials.

Create an outbound endpoint through the documented developer API and point it
at a public HTTPS ingress that forwards to this receiver. Do not expose this
sample directly without your normal TLS, firewall, availability, and durable
storage controls.

When the optional write values are present, startup calls the typed
`POST /api/v1/workspaces/{workspaceId}/stories` operation with
`Idempotency-Key`. Reusing that key with the exact same JSON request safely
returns the original story; changing the request while retaining the key is a
conflict. The operation accepts either a PAT or a service-account key with
`stories:write`; current credential state, workspace binding, and team
restrictions are revalidated by the API. Treat the idempotency key as durable
operation state, not as a timestamp or a random value regenerated on every
retry. The SDK transport will not retry the `POST` for you. Safe reads do retry
bounded `429` and `503` responses according to `Retry-After` and the SDK retry
policy.
