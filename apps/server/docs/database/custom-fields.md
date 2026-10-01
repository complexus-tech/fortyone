# Typed custom fields

`internal/modules/customfields` owns team definitions, retained select identities,
story values, their append-only audit writes, and current-value reports. Migration
`000200` adds this slice. Apply it through the existing release migration process
before activating the new UI; development tests use isolated testkit databases.

Definitions support text, number, money, date, select, and person. Types are
immutable. Money has one immutable ISO currency per definition; reports aggregate
one definition at a time and never combine currencies. Numeric values enter and
leave the API as decimal strings and are stored as PostgreSQL `NUMERIC`, with no
floating point conversion. Inputs accept up to 38 integral and 18 fractional
digits. `null` clears a value; zero and empty text remain real values.

Migration `000209` adds an optional definition `icon`. The API always returns a
catalog key or `null`; null uses the automatic type icon. Create may omit it.
Update omission preserves the current icon, while explicit null resets it. Keys
are `text`, `number`, `calendar`, `list`, `person`, `team`, `workspace`, `goal`,
`star`, `checklist`, `link`, `email`, `clock`, `work`, `attachment`, and `globe`.
Migration `000210` expands this stable catalog to 44 icons, adding `tag`, `pin`,
`objective`, `strategy`, `roadmap`, `home`, `health`, `approval`, `chat`, `comment`,
`share`, `image`, `video`, `microphone`, `book`, `help`, `info`, `analytics`,
`dashboard`, `workflow`, `kanban`, `sprint`, `automation`, `history`, `lock`, `key`,
`code`, and `warning`. The Projects picker groups these choices and searches
labels and synonyms. Deploy compatible API and Projects clients before selecting
expanded keys; older Projects schemas only understand the original catalog.
Migration `000211` adds 17 business and finance keys for a 61-icon catalog:
`money`, `coins`, `wallet`, `credit-card`, `bank`, `invoice`, `percent`,
`calculator`, `piggy-bank`, `target-money`, `building`, `megaphone`, `store`,
`package`, `phone`, `handshake`, and `shopping-cart`. Automatic money icons now
use the banknote glyph while stored `null` remains automatic. The original
`lock` key retains its identity and uses the corrected padlock glyph. New glyphs
use Hugeicons Stroke Rounded SVGs retained in the shared icon package under MIT.
Definitions in work backups retain the icon; older backups can omit it. Definition
create, update and archive write immutable scoped audit facts in the same
transaction, including previous and current icons. Value history is unchanged.

Definition changes require a live workspace administrator with access to the
team. Story values require a live member/admin, a current story/team binding, and
an unarchived story. Guests may read. Person values must identify an active
workspace and team member. Select choices must belong to that field and be
active when newly chosen. Fields and removed options are archived rather than
deleted, so existing labels and values remain available to read.

Value patches lock the team, current actor membership and story row. Optional
`expectedVersion` rejects stale edits. Every changed value and its old/new audit
record commit with a single story custom-field version; failed validation rolls
back the entire patch. Creation uses a bootstrap-bound, callback-scoped
transaction participant inside the existing serializable story transaction.
Custom values, creation activity, story sequence and outbox commit together.
Retried story creation retains the existing story-idempotency behavior.

Reports authorize the actor and field's team before running the aggregate in a
consistent read transaction. They include current, non-draft, unarchived and
undeleted stories. Created/completed date bases use UTC calendar dates; a date
field basis must belong to the same team. Missing and valued counts are explicit.
Count works for every field type and has no currency; numeric aggregates require
number/money. Average uses PostgreSQL's numeric division precision. Empty
average/min/max groups return an empty decimal string. These are current-state
reports, not a historical revenue ledger or completed-date value snapshots.

Definitions are bounded at 50 active and 200 retained fields per team, 100
active and 500 retained choices per field. Reports reject more than 500 groups
instead of returning a silently truncated result. The bounded batch story-value
endpoint accepts at most 100 story IDs, authorizes each story's team in one query,
and omits inaccessible stories. Fetch definitions once per team to render list or
board custom columns without issuing a request per story.

Verification: domain and transport tests cover exact decimals, nulls, dates,
immutable currency/type, unknown fields and structural input limits. PostgreSQL
integration tests cover current role/team fences, audit rollback, version guards,
retained options/fields, date-based exact aggregates, batch privacy, and atomic
story/outbox rollback.
