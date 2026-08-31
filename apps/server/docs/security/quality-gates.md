# API quality and security gates

Every pull request that changes `apps/server` runs one cancellable fast gate:

- `server-quality.yml` checks module integrity, gofmt, `go vet`, unit and
  architecture tests, committed-secret detection, and changed generated
  contracts;
- SQLC, OpenAPI, migration, and workflow checks run on the pull request only
  when their owning inputs change;
- superseded pull-request runs are cancelled.

`weekly-assurance.yml` and its manual dispatch run Staticcheck, the governed
Go security scan, reachable-vulnerability analysis, race and fuzz suites, full
generated-contract checks, disposable PostgreSQL migrations, SQLC vet, tagged
database and Redis tests, and the complete Projects quality suite.

`codeql-weekly.yml` is the tracked weekly/manual advanced CodeQL configuration
for Go and JavaScript/TypeScript. After this workflow is merged, a repository
owner must switch **Settings → Advanced Security → Code Scanning** from Default
Setup to Advanced Setup. Until that settings change, GitHub's dynamic Default
Setup remains authoritative and can disable result uploads from the tracked
workflow. GitHub Code Quality is a separate dynamic workflow; if it is enabled,
review or disable its push and pull-request scans independently to preserve the
weekly-only cost model.

The production release relies on the protected pull-request gate rather than
repeating it. It emits BuildKit SBOM and provenance attestations, blocks high or
critical runtime-image vulnerabilities with an immutable Trivy image, and
submits the commit-tagged Docker Hub images to ECS.

## Required-check settings

Repository branch-protection and ruleset settings are not stored in this tree.
No always-on classifier is included. Before marking the new fast-check jobs as
required, remove any stale contexts for `Server Quality and Security`, `Server
SQLC`, or `Projects Quality`. GitHub leaves path-filtered required workflows
pending when their paths do not match, so scope requirements to their owning
paths when the ruleset supports it. Add a classifier only if the live settings
cannot express that scope.

Scheduled workflows execute from GitHub's default branch. Keep
`weekly-assurance.yml` and `codeql-weekly.yml` on that branch (or change the
default branch to the release branch) before relying on their cron schedules;
manual dispatch remains available after the workflows are present there.

## Pinned tools

| Tool        | Version                                 | Purpose                                                                                                                    |
| ----------- | --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Staticcheck | `v0.7.0`                                | Go 1.25-compatible correctness/static analysis                                                                             |
| govulncheck | `v1.7.0`                                | Reachable Go vulnerability analysis                                                                                        |
| gosec       | `v2.29.0`                               | AST/SSA security analysis of handwritten Go; generated SQLC/OpenAPI code is excluded and regenerated from reviewed sources |
| Gitleaks    | `v8.27.2`                               | Committed secret detection                                                                                                 |
| SQLC        | `v1.31.1`                               | Typed query generation and database-backed vet                                                                             |
| actionlint  | `v1.7.7`                                | Workflow syntax, expression, and job contract validation                                                                   |
| oasdiff     | `v1.28.0`                               | Pull-request comparison that blocks incompatible public OpenAPI changes                                                    |
| Trivy       | `v0.74.0` OCI digest `sha256:62b1e65e…` | Release-image operating-system and application vulnerability gate                                                          |

Go tools run with an exact module version. Gitleaks uses an official release
archive with an OS/architecture-specific SHA-256 allowlist. CI first feeds it a
known non-secret fixture and fails if the default rules do not detect it. This
self-test is required because a scanner that exits successfully without loading
effective rules is worse than an explicit failure.

Every third-party GitHub Action is pinned to a reviewed 40-character commit;
the adjacent comment records the human-readable release. The architecture suite
scans every workflow and rejects moving tags or branches. Dependabot or a
dedicated maintenance change may propose upgrades, but the commit, release
notes, permissions, inputs, and output behavior must be reviewed together.

The production release uses the repository's existing paired AWS access-key
secrets and `AWS_REGION`. The architecture regression test keeps those names
aligned with the deployed repository configuration. The associated IAM policy
must remain restricted to the ECS operations and task-definition roles used by
this workflow, and the key pair must be rotated operationally.

Do not replace a pin with `latest`. Upgrade one tool at a time, review release
notes and new findings, update the pin/checksum, run the full workflow, and keep
any new suppression narrow and documented.

The Trivy release gate has no blanket ignore file and does not ignore unfixed
high or critical findings. A temporary waiver requires a reviewed exact
vulnerability identifier, owner, impact analysis, compensating control, and
expiry; encode it narrowly rather than weakening the severity or exit-code
flags. The commit-SHA image tag is the release identity used by both ECS task
definitions.

## Governed G104 transport exception

`G104` (unchecked returned errors) is not blanket-suppressed. The complete
pinned gosec scan is emitted as JSON and passed to
`internal/tools/g104guard`. Every finding for every other rule remains
blocking. For `G104`, the guard parses the current Go syntax tree and rejects
every finding except a standalone four-argument call to `Respond` or
`RespondError` imported from the canonical
`github.com/complexus-tech/projects-api/pkg/web` package.

The narrow transport exception exists because those legacy handlers have
already committed or attempted the response, the shared response writer records
serialization/write failures on the request span, and a second HTTP response is
not a valid recovery. Rewriting hundreds of handlers in one mechanical change
would alter control flow without improving client recovery. New or deliberately
refactored handlers should return `web.Respond(...)` or
`web.RespondError(...)` directly so the normal handler boundary receives the
error.

The guard resolves each G104 scanner location against the AST and the canonical
import binding; matching text in a comment, a shadowed import, local lookalike
function, assignment, deferred call, cleanup function, logger, cache, database,
compression, or credential path is not allowed. Malformed reports, invalid rule
identifiers, duplicate locations, package-analysis errors, zero-file scans,
paths outside the server root, and stale scanner locations all fail closed. Its
output contains only rule identifiers, repository-relative locations, and
counts, never gosec source snippets or raw diagnostics.

All other unchecked errors must be handled locally with the correct operational
semantics. A `#nosec` annotation must name the rule and explain the validated
invariant; a bare suppression is rejected in review.

## Secret-scan scope

Local scanning materializes the Git index into a private temporary directory and
scans that snapshot. This prevents ignored developer `.env` files from entering
scanner output while matching CI, where the pull-request commit is the index.
Stage a new file before relying on the local result. Findings are always
redacted; raw secret output must never be enabled.

## Commands

```bash
make bootstrap-tools
make tool-versions
make check-fast
make quality-check
make test
make test-race
make test-fuzz
make g104-check
make security-check
make gitleaks-check
make generated-check
OPENAPI_BASE_REF='<reviewed base commit>' make openapi-breaking-check
make ci
```

`make generate` is the intentional writer for SQLC, configuration, inventory,
and migration documentation. It deliberately excludes the architecture debt
baseline, whose separate command requires review. `make check-fast` and
`make ci` are read-only verification paths; neither accepts generated drift or
rewrites a baseline. The fuzz target list is explicit and time-bounded so a new
fuzzer joins the governed surface through a visible review.

PostgreSQL-tagged tests and `make sqlc-vet` additionally require the explicit
disposable database contract described in `docs/database/sqlc.md`.
