# Notification authorization

Notification IDs and queued task payloads are untrusted references. Possessing
an ID does not authorize an inbox read, mutation, portal read, or email send.
Every persistence operation rechecks current actor and resource state.

## Authorization matrix

| Operation                     |             Active account | Live workspace |                   Workspace membership |                         Team/resource scope |                                Portal contributor scope |
| ----------------------------- | -------------------------: | -------------: | -------------------------------------: | ------------------------------------------: | ------------------------------------------------------: |
| create workspace notification |        actor and recipient |       required | required for human actor and recipient | required unless system actor policy applies |                                          not applicable |
| list/count workspace inbox    |            recipient actor |       required |                     admin/member/guest |                       current resource team |                                       feedback excluded |
| read/unread/delete one        |            recipient actor |       required |                     admin/member/guest |                       current resource team |                                       feedback excluded |
| read/delete workspace batch   |            recipient actor |       required |                     admin/member/guest |                          each row rechecked |                                       feedback excluded |
| get/update preferences        |            recipient actor |       required |                     admin/member/guest |                              not applicable |                                          not applicable |
| list/count/read portal inbox  |            recipient actor |       required |                           not required |                          live feedback item | public portal plus active unblocked account contributor |
| read pending email/digest     |                  recipient |       required |              required for non-feedback |                       current resource team | public portal plus active unblocked account contributor |
| key-result audience           | event actor and recipients |       required |                               required |                              objective team |                                          not applicable |

Admin membership can see workspace resources across teams where the existing
product policy allows it. Member and guest roles require an explicit current
team membership. Guest is a supported workspace role, not a bypass: removing
the guest from the resource team immediately hides stale notifications.

## Actor and recipient rules on create

Creation validates two distinct identities:

- the actor is who caused the event;
- the recipient is who may receive it.

For a human workspace event, the actor must be active and currently authorized
to the exact workspace team/resource. The recipient must independently be
active and currently authorized. This prevents an event from another tenant
from manufacturing a notification that appears to come from a valid local
user.

System users are the explicit exception for background behavior such as Maya,
scheduled reminders, and the proxy actor used for a verified guest feedback
comment. A system actor must still be an active system user. It does not grant
the recipient access: the workspace or portal recipient/resource predicates
remain mandatory.

For feedback caused by an account user, the human actor may be either a current
workspace/team member acting as product staff or an active unblocked account
contributor to the exact public portal. The system guest proxy covers guest
comments whose contributor is not a `users` row.

## Revocation behavior

Authorization is based on current database state, never only on middleware or
the state that existed when the event was published.

- disabling the recipient account denies inbox, preference, portal, and email
  delivery reads;
- removing workspace membership denies the internal inbox and preferences;
- removing team membership hides notifications for that team's story,
  comment, objective, or key result;
- deleting a story or feedback item hides its stale notification;
- blocking an account contributor or making the portal private denies the
  portal inbox and pending feedback email; and
- deleting a workspace denies all module reads and mutations.

Single-resource workspace mutations return not found after current workspace
authorization succeeds but the notification is missing, owned by another
recipient/tenant, or no longer resource-visible. This avoids disclosing whether
another user's notification ID exists. A caller without workspace or portal
authorization receives forbidden. Dedupe-content disagreement returns
conflict. Invalid finite types, UUIDs, patches, and pagination return invalid.

HTTP maps those categories to 404, 403, 409, and 400 respectively. Unexpected
database errors remain 500 responses and must not expose SQL or credentials.

## Structured-message privacy

Notification `message` is JSONB, but it is a typed domain model, not arbitrary
API output. Strategy snapshots can include internal objective and key-result
detail. `Notification.Public()` removes the snapshot and replaces its template
and variables before HTTP or realtime serialization.

New structured fields require an explicit public mapping decision and tests.
Never return raw JSONB or a generated SQLC row from a handler.

## Security review checklist

Before adding or changing a notification path, verify all of the following:

1. actor, recipient, workspace/portal, and resource IDs come from trusted event
   or authenticated context rather than a client-controlled override;
2. the SQL statement checks active identity and current authorization for every
   referenced resource;
3. a system actor is explicitly required and cannot widen recipient scope;
4. dedupe keys include immutable event identity plus recipient identity;
5. workspace and portal inboxes cannot show each other's notification types;
6. queued delivery reads repeat authorization and preference checks;
7. mutations use finite typed intent rather than a field map or SQL fragment;
8. ordering has a unique tie-breaker and all integer conversions are checked;
9. logs and errors omit message contents, bearer values, emails, and SQL; and
10. PostgreSQL 18 tests include cross-tenant, inactive, revoked, guest, blocked,
    private-portal, concurrent replay/mutation, and rollback cases.

See [Notification persistence and delivery](../database/notifications.md) for
query ownership, idempotency, recovery, and index details.
