# Actor and authorization model

Authentication answers **who and which credential made the request**. Authorization answers **whether that actor may perform this operation on this current resource**. FortyOne keeps those questions separate so an integration or service account is never recorded as the human who installed it.

The executable contracts live in:

- `internal/platform/auth`: immutable actor, principal kinds, scopes, team restrictions, and context propagation;
- `internal/platform/authorization`: workspace role hierarchy and service policy decisions;
- `internal/platform/http/middleware`: credential authentication, workspace selection, and cheap route-level scope rejection.

## Actor contract

Every authenticated operation carries an `auth.Actor` with:

| Field          | Meaning                                                                                           |
| -------------- | ------------------------------------------------------------------------------------------------- |
| `PrincipalID`  | The human, service account, application, system, or external principal that acted.                |
| `Kind`         | The explicit principal kind. New kinds are denied by policies until each policy opts in.          |
| `WorkspaceID`  | The selected tenant after current membership has been resolved.                                   |
| `CredentialID` | The session, token, key, or grant used, when it has a durable identifier.                         |
| `Scopes`       | Credential permissions that narrow product access. A scope never creates membership or ownership. |
| `TeamAccess`   | Optional credential-level team restriction. Product team membership remains authoritative.        |

Browser sessions become `human_user` actors with the internal `first_party:*`
scope. PAT and service-account verification now constructs explicit
`personal_token` and `service_account` actors. The strict machine middleware is
installed only on the documented `/api/v1` route group and is not installed on
legacy browser routes. Preview resource reads and webhook management allow PAT
actors only; service-account authorization remains future work. OAuth
authentication remains future work.

Actor values and their scope/team collections are defensively copied when they enter or leave context. Do not add provider SDK tokens, raw keys, request bodies, or mutable application models to an actor.

## Decision sequence

For a workspace operation, apply all of these checks:

1. authenticate the credential and construct an actor;
2. resolve the current active workspace membership from PostgreSQL and bind the workspace to the actor;
3. reject a missing coarse route scope in middleware;
4. load the resource through a workspace-scoped repository query;
5. call the service policy with the explicit allowed principal kinds, required scopes, current workspace role, team/resource state, and requested change;
6. persist with the workspace predicate repeated in SQL;
7. record the actor, credential/installation, operation, resource, request or delivery ID, decision code, and safe metadata in the audit event.

Workspace membership and role are deliberately not served from the general Redis cache. This makes removal and demotion effective on the next request even if cache invalidation or Redis is unavailable. If authorization caching is reintroduced, it requires a separate ADR and a version/generation design that proves immediate revocation.

First-party system actor UUIDs are also resolved from live PostgreSQL state at
process startup rather than Redis. Only an active `is_system` user matching the
closed application key catalog is accepted. See
[First-party system actor resolution](system-actors.md).

Workspace middleware receives only a current-membership resolver capability; it
cannot query through a handler-owned database handle. Successful resolution
also starts a best-effort access-metadata write with a 250 ms upper bound. That
write repeats current active membership in SQL and cannot extend or recreate
authority. Its failure is logged and does not fail an otherwise authorized
business request.

## Current scope catalog

The supported public-shaped scopes are:

```text
workspaces:read
teams:read
stories:read
stories:write
comments:read
comments:write
labels:read
labels:write
sprints:read
objectives:read
objectives:write
webhooks:manage
integrations:manage
service_accounts:manage
```

`first_party:*` is the internal capability marker assigned only to authenticated
first-party human actors. It must not be issued as a PAT, API key, OAuth grant,
or provider credential.

## Adding a protected operation

1. Reuse an existing scope or add one to the catalog with an endpoint mapping and compatibility review.
2. Add `mid.RequireScopes(...)` at the route for early credential rejection.
3. Define the service policy input after the resource is loaded. List allowed principal kinds explicitly.
4. Repeat workspace and ownership predicates in the SQLC query.
5. Add a matrix covering principal kind, scope, workspace role, team restriction, same-tenant and cross-tenant IDs, resource visibility, and mutation-specific rules.
6. Map denials to the stable external error envelope without exposing whether a cross-tenant resource exists.

Never authorize only from an ID in a URL, only from a credential scope, only from middleware, or only from a cached role. Administrative scope and workspace admin role are separate requirements; neither implies the other.

## Test standard

Policy tests are pure table-driven tests. Repository contract tests use two workspaces and prove every read/write predicate is tenant-scoped. HTTP tests prove missing authentication, missing scope, and policy denial mappings. Revocation tests change membership or credential state and make the next request without sleeps or cache flushes.
