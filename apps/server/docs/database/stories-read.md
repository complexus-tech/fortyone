# Stories read persistence and visibility contract

Actor-facing story reads use module-owned SQLC queries and the native pgx/v5
pool. Primary story mutations plus bulk lifecycle, label, collaborator,
watching, comment-read, and schedule-transition capabilities now use the same
typed boundary. There is one production stories repository and no fallback
persistence path; related write, automation, retention, and deletion-outbox
contracts are documented in
[Story mutation persistence](stories-mutations.md).

## Capability and source map

| Read capability                                            | SQLC query                                                                        | Handwritten adapter                             |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------- | ----------------------------------------------- |
| Story by ID                                                | `GetVisibleStory`                                                                 | `read_single.go`                                |
| Story ID and projection by `TEAM-123` reference            | `GetVisibleStoryIDByRef`, then `GetVisibleStory`                                  | `read_single.go`                                |
| Stories assigned to, reported by, or shared with the actor | `ListMyVisibleStories`                                                            | `read_lists.go`                                 |
| General typed story list                                   | `ListVisibleFilteredStoryRows` in page mode                                       | `read_filtered.go`                              |
| Visible active-story count                                 | `CountVisibleStories`                                                             | `read_filtered.go`                              |
| Initial grouped story columns                              | `ListVisibleStoryGroupCatalog` and `ListVisibleFilteredStoryRows` in grouped mode | `read_filtered.go`                              |
| One subsequent group page                                  | `ListVisibleFilteredStoryRows` in page mode with a typed group key                | `read_filtered.go`                              |
| Team stories in one status category                        | `ListVisibleStoriesByCategory`                                                    | `read_lists.go`                                 |
| Batched child-story projection                             | `ListVisibleSubStories`                                                           | `read_lists.go`                                 |
| Visible related-story projection                           | `ListVisibleStoryAssociations`                                                    | `read_single.go`, `read_association_mapping.go` |
| Paginated visible comment roots and batched replies        | `ListVisibleStoryCommentRoots`, `ListVisibleStoryCommentReplies`                  | `comment_reads.go`                              |
| Visible comment or reply by comment and story ID           | `GetVisibleStoryComment`                                                          | `comment_reads.go`                              |

Reviewed SQL lives in:

- `internal/modules/stories/repository/queries/stories.sql` for single,
  personal, category, child, and association reads; and
- `internal/modules/stories/repository/queries/filtered_stories.sql` for the
  typed list, count, group catalog, grouped window, and group-page surface; and
- `internal/modules/stories/repository/queries/comment_reads.sql` for the
  tenant-scoped comment tree and parent-notification lookup.

Generated pgx code lives in
`internal/modules/stories/repository/sqlc` and must never be edited by hand.
Adapters map generated rows into dependency-neutral contracts in
`internal/modules/stories/domain`; generated parameter, row, and database-null
types never escape the repository.

The service-facing aliases in `service/models.go` preserve the existing API
shape while keeping dependency direction clear:

```text
HTTP or internal caller
  -> stories service (actor and scope policy)
  -> stories domain (typed query and explicit read scope)
  -> stories repository (validation, SQLC parameters, row mapping)
  -> PostgreSQL (live authority proof and data selection)
```

## Actor, workspace, and team authority

Every migrated repository method receives an explicit `ReadScope`. The
repository never extracts identity from `context.Context`. The scope contains:

- the authenticated actor ID;
- the selected workspace ID;
- whether the credential has unrestricted team access; and
- the credential's exact allowed-team IDs when access is restricted.

The service constructs this scope from the platform actor, requires
`stories:read`, and rejects an actor bound to a different workspace. Each SQL
statement then rechecks mutable authority at execution time:

1. the actor still exists and is active;
2. the actor still belongs to the selected workspace;
3. the actor still belongs to every returned story's team;
4. the story and team belong to the same workspace; and
5. credential-level team restrictions include the story team, unless the
   credential is explicitly unrestricted.

Request filters never supply actor authority. The compatibility
`CurrentUserID` and `WorkspaceID` fields are cleared by repository
normalization; only `ReadScope` is authoritative.

Missing, cross-workspace, inaccessible-team, and inactive-actor single-story
reads all return `domain.ErrNotFound`. This avoids disclosing that a hidden
story exists. List, count, and group reads return no visible data under the same
authority failure. Workspace middleware is useful routing context, but it is not
the repository's authorization source of truth.

Actor-scoped list responses are not stored in the application cache. A cache
hit would bypass the live workspace/team membership proof and could retain data
after membership or credential-team access is revoked. Reintroducing a read
cache requires an authoritative authorization-version component in the key and
an invalidation contract for membership and credential changes; principal ID
or workspace ID alone is insufficient.

Associations apply the visibility proof to both the requested story and the
related story. An association to an inaccessible team is omitted rather than
leaking its title or identifiers. Child-story loading repeats actor, workspace,
team, and credential predicates and runs as one batch for the visible parent
page.

## Typed filter contract

`domain.StoryFilters` is the only general list filter input. SQL parameters are
static and typed; there is no field-name interpolation, dynamic WHERE builder,
or `map[string]any` read path.

Supported filters are:

- inclusion and exclusion UUID lists for statuses, assignees, reporters, teams,
  sprints, and labels;
- collaborator inclusion;
- title/description contains and does-not-contain text;
- inclusion and exclusion lists for priority and estimate;
- status category inclusion;
- parent, objective, excluded objective, and key-result IDs;
- has-assignee, has-no-assignee, has-blocker, assigned-to-me,
  collaborating-with-me, and created-by-me flags;
- inclusive created, updated, start-date, deadline, and completed ranges;
- exact start-date and deadline exclusions;
- completed and not-completed flags;
- child-story inclusion; and
- archived and deleted modes.

Within one inclusion list, a row matches any supplied value. Different filter
fields combine with AND. Exclusions retain NULL values unless the corresponding
positive requirement excludes them. Multiple enabled personal flags combine
with OR, matching the existing “my relevant stories” behavior.

Nil pointers and empty slices mean “not supplied.” Boolean filters take effect
only when true. `IncludeArchived=true` selects only archived rows and
`IncludeDeleted=true` selects only deleted rows; false or omitted selects only
non-archived or non-deleted rows. Date boundaries are inclusive. Priority is one
of `Urgent`, `High`, `Medium`, `Low`, or `No Priority`; estimates are
`1`, `2`, `3`, `5`, or `8`; categories are `backlog`, `unstarted`,
`started`, `paused`, `completed`, or `cancelled`.

The current schema has no story epic column. An explicit epic filter is rejected
with `ErrInvalidReadQuery` rather than silently broadening the result.

The repository rejects zero UUIDs, more than 100 values in one filter, invalid
enum values, contradictory assignee/completion flags, reversed date ranges,
negative or excessive offsets, and unsupported limits. HTTP parsing also rejects
malformed explicitly supplied UUIDs, booleans, integers, arrays, and dates;
invalid input is never discarded and converted into an unfiltered query.

## Grain, ordering, grouping, pagination, and nulls

A single-story row has one story as its grain. List rows also have one parent or
requested story as their grain. Labels, collaborators, and notification audience
IDs are typed `uuid[]` projections with deterministic ordering. The repository
exposes non-nil empty slices when no values exist. Children and associations use
separate bounded queries rather than nested JSON decoding or one query per
parent.

`GetVisibleStory` preserves the existing direct-ID lifecycle behavior and can
return an archived or deleted row to an authorized actor. Reference lookup
excludes deleted rows. General, category, grouped, and count reads exclude
archived and deleted rows unless the typed general filter explicitly selects
the corresponding lifecycle mode. The personal list preserves its legacy
contract: it excludes deleted rows but can include archived rows. Child and
association projections likewise preserve their lifecycle compatibility and
must not be treated as independent active-story lists.

The ordinary `/stories` list returns top-level rows by default, ordered by
`created_at DESC, id ASC`, and capped at 1,000 rows. Trusted internal callers
may request an explicit positive limit up to 1,001 and a bounded offset; HTTP
query parameters cannot override those internal limits.

`ListMyVisibleStories` returns top-level rows for which the actor is the
assignee, reporter, or a collaborator. It is ordered by
`created_at DESC, id DESC` and capped at 1,000 rows.

Grouped reads accept these enums:

| Dimension | Values                                                     |
| --------- | ---------------------------------------------------------- |
| group     | `status`, `assignee`, `priority`, `team`, `sprint`, `none` |
| order     | `created`, `updated`, `priority`, `deadline`, `completed`  |
| direction | `asc`, `desc`                                              |

Initial grouped reads return at most 100 stories per group and retain visible
empty catalog columns such as statuses, teams, priorities, assignees, and
sprints. Catalog construction is itself actor/team scoped and is rejected when
it exceeds 1,000 entries, before the story window query runs. A subsequent
group page requires a validated group key: a UUID for status/team, a UUID or
`null` for assignee/sprint, a valid priority, or `none` for the ungrouped view.

All dynamic orders end with `created_at DESC, id ASC` as deterministic
tie-breakers. Group pagination uses page numbers for compatibility:

- page: 1 through 10,000;
- page size or stories per group: 1 through 100; and
- SQL limit: requested size plus one, so `hasMore` needs no extra count.

The count endpoint counts visible, active-lifecycle stories across both parent
and child rows. The ungrouped grouped view carries a windowed total count for
the matching typed filters. Offset pages are stable for an unchanged dataset
but are not snapshots: concurrent inserts or updates can shift later pages. A
versioned public API should eventually replace the compatibility page number
with an opaque keyset cursor before removing the existing contract.

Optional database columns map to pointers. A missing objective or sprint keeps
both its ID and embedded summary nullable. Priority and estimation scheme are
normalized in SQL to `No Priority` and `tshirt`. Estimate labels are derived
in memory from the typed scheme selected by SQLC; the repository does not issue
a follow-up estimation query.

## Index and plan budget

The schema provides the primary access paths used by this slice:

- `stories_pkey` for story-by-ID;
- `unique_team_sequence (team_id, sequence_id)` for reference lookup;
- `idx_stories_search_workspace_team` for live workspace/team rows;
- `idx_stories_workspace_team_deleted_parent` for team pages and children;
- `idx_stories_status_id`, `idx_stories_sprint_id`,
  `idx_stories_key_result_id`, and
  `idx_stories_workspace_team_estimate_unit` for typed filters; and
- `idx_story_comments_roots_page` and
  `idx_story_comments_replies_page`, introduced by reversible migration 000166,
  for bounded root pagination and ordered batched reply hydration; and
- primary or unique keys on users, workspace memberships, team memberships,
  labels, collaborators, watchers, and associations.

PostgreSQL 18 integration tests exercise the real migration chain through the
current repository head, move the isolated test database to 000166, prove the
000166 down/up transition, restore the head, and assert the two comment access
paths select their intended partial indexes. Before production rollout, capture
`EXPLAIN (ANALYZE, BUFFERS, SETTINGS, FORMAT JSON)` on a staging snapshot with
representative tenant and team cardinality. Record the PostgreSQL settings,
cardinalities, chosen filter parameters, and resulting plan with release
evidence; never commit production values or data.

| Operation                  | Warm-cache p95 budget |                      Maximum returned rows | Plan guardrail                                                                                            |
| -------------------------- | --------------------: | -----------------------------------------: | --------------------------------------------------------------------------------------------------------- |
| get by ID or reference     |                 25 ms |           1 story plus bounded projections | indexed story lookup; no whole-story-table scan                                                           |
| ordinary or personal list  |                100 ms |                              1,000 parents | workspace/team index path; bounded result and in-memory mapping                                           |
| filtered group window      |                150 ms | 101 rows per visible group before trimming | authority/filter predicates before WindowAgg; no disk spill                                               |
| one group or category page |                 75 ms |                   101 rows before trimming | filter before bounded sort; no per-parent query loop                                                      |
| visible count              |                 75 ms |                                   1 scalar | workspace/team authority predicates use indexed joins                                                     |
| children or associations   |                 50 ms |                 rows for one returned page | one batched query; no N+1                                                                                 |
| comment roots plus replies |                 75 ms |        101 roots plus their direct replies | `idx_story_comments_roots_page`, then one `idx_story_comments_replies_page` query; no per-root query loop |

Treat an unexpected whole-table sequential scan of `stories`, a disk-spilling
sort/window, more than 5,000 shared-buffer reads for one page, or a twofold
regression from an accepted staging plan as a release blocker. A sequential
scan can be legitimate only for a deliberately tiny fixture; it is not
acceptable evidence for representative production cardinality.

## Production boundary

The production repository constructor is explicit:

```go
storiesrepository.New(log, pool)
```

The pgx pool cannot be silently omitted from production wiring. Stories and
comments expose caller-owned typed ports so services and focused unit tests stay
independent of generated persistence types. Production always narrows the one
pgx adapter to those ports; test doubles do not create another production
storage implementation. The complete mutation and persistence boundary is
documented in
[Story mutation persistence](stories-mutations.md).

The former raw general list, count, grouped list, group-page, personal list,
category list, reference query, and comment-tree paths have been deleted. Do
not add another read query builder or reintroduce SQLx as a production fallback.
Test-only doubles may isolate service behavior; production typed reads require
the scoped SQLC capability.

## Adding or changing a story read

1. Add a statically named SQLC query to the capability-appropriate file under
   `repository/queries`.
2. Bind all request input through typed SQLC parameters. Never interpolate a
   field, enum, direction, group key, actor, workspace, or team identifier.
3. Repeat active actor, current workspace membership, team membership,
   story/team workspace binding, and credential-team predicates in SQL.
4. Add or extend a domain-neutral input/output contract in `stories/domain`;
   generated SQLC types stay inside the repository.
5. Validate enums, bounds, contradictions, and scalar presence before invoking
   SQLC.
6. Regenerate with `make sqlc-generate`; never edit `repository/sqlc` by hand.
7. Map rows in a cohesive capability file; do not recreate the deleted generic
   `repository/queries.go` or grow `service/stories.go` with another use case.
8. Add unit tests and PostgreSQL 18 cases for visible, cross-tenant,
   inaccessible-team, restricted-credential, and inactive-actor behavior.
9. Capture and compare a representative query plan against the budgets above.

## Verification

Run from `apps/server`:

```bash
make sqlc-generate
make sqlc-check
go test ./internal/modules/stories/...
go test -count=1 -race ./internal/modules/stories/...
go vet ./internal/modules/stories/...
make architecture-check

TEST_DATABASE_URL='postgresql://test-role:password@127.0.0.1:5432/postgres?sslmode=disable' \
  go test -count=1 -race -parallel=1 -tags=integration \
  ./internal/modules/stories/repository
```

The integration role must target a disposable PostgreSQL 18 server and be able
to create and drop isolated test databases. The suite applies the real migration
chain. In particular,
`TestFilteredGroupedAndCountReadsShareVisibilityBoundary` proves that ordinary
lists, counts, grouped columns, group pages, inactive actors, and
credential-restricted actors share the same visibility boundary.
`TestStoryCommentReadsPreserveTreeShapeAndLiveTenantVisibility` proves the
comment root/reply shape, parent lookup, workspace isolation, credential team
restriction, and immediate membership revocation behavior.
`TestStoryDuplicationIsTenantFencedAtomicConcurrentAndPlanBacked` proves
source-version CAS, live membership, cross-tenant denial, event-conflict
rollback, concurrent sequence allocation, and the indexed source lookup.
`TestStoryAssociationMutationUsesCASAndRollsBackEventsAtomically` proves
tenant isolation, rollback, and concurrent association compare-and-swap.
