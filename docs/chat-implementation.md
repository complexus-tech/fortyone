# Chat implementation runbook

## Implemented behavior

- Team conversations and private, one-to-one Direct Messages use persisted messages, cursor pagination, threads, reactions, read state, search, and Files & links. The web timeline is virtualized; editing drafts survive offscreen row unmounts.
- Messages support rich text, structured mentions, attachments, author editing, and deletion. Deleted thread roots retain a tombstone when replies remain.
- Writes use authenticated HTTP endpoints. Real-time updates use the existing user-scoped SSE connection and Redis publishing. No separate WebSocket service is required. Updates are cache invalidation hints; clients fetch authorized data from the API.
- Message changes and delivery events commit together. The worker drains a durable outbox, with immediate wake-ups and periodic recovery. Direct Messages, mentions, and replies can notify eligible users; ordinary team traffic does not notify every member. Delivery rechecks current access and deduplicates notifications.
- Attachment upload/download uses the existing storage lifecycle and short-lived authorized download URLs. Scan-status fields exist, but the current direct-upload path permits unscanned files; an active scanning worker has not been established by the reviewed implementation. Removing attachments during edits, message deletion, or pending-upload cancellation retires unreferenced metadata and queues storage deletion in the same transaction. Team/account deletion uses the same shared-reference safeguards. Account erasure preserves anonymous team history and removes private Direct Messages.
- Abandoned pending Chat uploads expire after 24 hours. The existing minute-cadence attachment cleanup processes at most 100 expired associations per run; sent messages and files still referenced elsewhere are preserved. Resuming an older unsent draft may require reattaching its files.
- Maya reuses the existing invocation runner, model provider, usage limits, and task tools. A structured `@Maya` mention in a team conversation starts a real run; a literal name in ordinary text does not. Direct Messages do not invoke Maya. Context is bounded to the relevant conversation/thread, and requester access plus source revision are rechecked during execution.
- Maya can answer questions, summarize discussion, and propose tasks for the mentioned team. Task creation requires explicit approval by the original requester. Saved proposals, execution receipts, deterministic reply IDs, and real task references prevent duplicate work and distinguish proposals from completed actions.
- Human typing hints use short-lived, scoped SSE events rather than stored messages. The browser bounds outgoing hints, expires inactive people, and clears presence on disconnection. Maya's active invocation appears as a muted “Maya is thinking…” indicator below the composer. Created tasks reuse the assistant's shared task-results renderer.
- Plain and formatted messages automatically recognize safe web/email links; code stays literal. Shared links are also extracted into Files & links without fetching arbitrary remote pages.

## Local startup

Run these commands from `apps/server`:

```sh
make migrate-version
make migrate-up
make dev
```

`make dev` runs the API with Air. Use `make develop` to run without live reload. In a separate terminal, start the worker:

```sh
make worker
```

From the repository root, start the web application:

```sh
pnpm --filter projects dev
```

The local schema checkpoint is **232**: migration 229 adds team Chat, 230 adds Chat notification types, 231 adds voice-note metadata and durable audio transcriptions, and the concurrently delivered 232 adds Maya reply-seen notification state. Migrations 231 and 232 have been applied to the local database; this is not a production migration record. Do not edit applied migrations; use a new migration for subsequent changes. The API and worker must both run the updated code. Restart an already-running worker after changing its composition or handlers.

## Local sample conversations

From `apps/server`, seed the existing local **Second** workspace:

```sh
python3 scripts/seed-chat.py --workspace second
```

The sample dataset contains **180 root messages, 60 replies, and 12 messages in Direct Messages**, providing history for scrolling and thread testing. Seeding does **not** enqueue outgoing historical notifications or Maya jobs. Treat these records as local sample content, not actual team activity. Run the seeder only against the intended local database.

## Configuration and dependencies

- Use the server's existing local environment/configuration loading. PostgreSQL and Redis must be available to both processes; there is no supported full-stack Compose setup.
- Keep the existing storage provider, attachment/profile buckets, credentials, and scanning configuration. API and worker must agree on the attachment storage route so queued deletion jobs reach the correct provider/container. No Chat-specific bucket is required.
- Email uses the existing SMTP/notification configuration. The current local SMTP endpoint is `127.0.0.1:1025`; keep local tests on the local mail service. Push and email remain subject to existing delivery configuration and user preferences.
- Maya uses the existing AI API key/model, workspace entitlement, rate limiter, daily usage budget, and mutation-confirmation configuration. Never replace missing production configuration with a fake model. Deterministic fake providers are confined to automated tests.
- The existing Second workspace trial had expired. Its local database trial end was extended by seven days on October 16, 2026, solely for the real Maya smoke test. External billing and production entitlement were not changed.
- Storage credentials, SMTP credentials, model keys, and database URLs belong in the established secret/configuration mechanism, never this document or test output.

## Validation checkpoint

Targeted API/worker compilation, unit tests, and real PostgreSQL integration suites passed for Chat, Maya invocations, attachments, team deletion, and account erasure. Database tests use isolated databases, not the seeded workspace. Coverage includes atomic message/run acceptance, requester-only approval, actual task receipts, idempotent replies, access boundaries, attachment retirement, shared-file preservation, and rollback on failed retirement.

Useful commands from `apps/server`:

```sh
go test ./internal/bootstrap/api ./internal/bootstrap/worker ./internal/modules/chat/...

# Requires TEST_DATABASE_URL pointing to an approved local PostgreSQL instance.
# The test harness creates and drops isolated databases.
go test -tags=integration -p 2 \
  ./internal/bootstrap/chatadapter ./internal/bootstrap/agentadapter \
  ./internal/modules/chat/... ./internal/modules/attachments/repository \
  ./internal/modules/teams/uow ./internal/modules/users/uow -count=1
```

These automated results verify database behavior and queued delivery/retirement. They do **not** establish successful cloud-object deletion, external notification delivery, production readiness, or sustained concurrency.

### Browser acceptance on October 16, 2026

The real application at `http://localhost:3000/second/chat` was checked separately from the UI-only preview at `http://127.0.0.1:4177/?theme=dark`:

- A message persisted after refresh; its reaction and edited thread reply were displayed correctly.
- A newly sent message appeared in a second already-open tab through SSE without reloading.
- A structured `@Maya` mention produced a real model-backed summary of the recent conversation. A second request proposed one assigned task; approving it created **TM-267, Verify chat keyboard navigation**, and displayed the shared assistant task-results UI in the thread.
- A small test attachment uploaded as a file-only message and its authorized download opened the exact uploaded text. Files & links displayed this file plus the extracted links from a newly sent message.
- Both `https://example.com/guide` and `www.example.com` rendered as clickable links in the conversation.
- Conversation search returned matching messages and replies; opening a reply result navigated into its thread. A further real Maya request displayed “Maya is thinking…” beneath the composer before the answer appeared.
- The larger reply heading, composer padding, centered supplied send icon, supplied new-chat icon, stacked reply avatars, and removal of the final code-block toolbar control were included in the refreshed UI preview.

In-app notification badges updated after Maya replies. Local Mailpit capture is running on SMTP `127.0.0.1:1025` and UI `127.0.0.1:8025`, without outbound relay; delivery into it has not yet been verified. Native desktop notification display, external email/push delivery, and storage-object deletion remain separate deployment checks.

### Final automated checks

- Chat frontend: 30 tests passed, scoped ESLint passed, and the full Projects TypeScript check passed. ReactDoctor reported no errors and five advisory warnings.
- Shared SSE contracts/connection, desktop notification mapping, and notification destinations: 56 tests passed. Native notification-routing unit tests: 7 passed.
- SQLC generation/compilation checks, API inventory checks, migration validation, scoped Go vet, and the targeted Go/PostgreSQL suites passed. Deep-history database checks exercised 10,100 messages and verified indexed cursor plans; this is not a concurrent-user load test.
- The repository architecture-debt baseline check still fails on pre-existing, untouched desktop cross-module imports and oversized Maya, notifications, and users test files. The baseline was not relaxed to hide these failures.

## Deployment and scale checks still required

The initial planning target is **1,000 concurrent users**, not a measured capacity guarantee. Before launch and before increasing toward 5,000, 10,000, or 50,000 concurrent users:

1. Load-test representative room sizes, message/attachment rates, thread depth, notification fan-out, and Maya demand. Concurrent connections alone do not predict cost or throughput.
2. Verify proxy/load-balancer SSE buffering, idle timeouts, HTTP/2 behavior, reconnects, and cross-replica delivery. Measure connections, memory, Redis traffic, event lag, and client refetch volume.
3. Measure database query plans and latency for long histories, search, unread counts, membership checks, and attachment cleanup. Set connection-pool limits against the actual PostgreSQL capacity.
4. Exercise worker restart, duplicate delivery, Redis failure, database retry, revoked access, and storage deletion failure. Monitor outbox age, retries, leases, queue depth, and dead/stalled work; establish retention and recovery procedures.
5. Validate production scanning/download access, notification preferences, unsubscribe behavior, SMTP/push delivery, and task approval auditability using the actual deployment configuration.
6. Budget separately for API/SSE compute, database and Redis capacity, storage/egress, notification delivery, and model tokens. Enforce Maya quotas and provider limits, and measure cost per active workspace before expanding capacity.

Scale workers, API replicas, connection capacity, and database resources from those measurements. The current local checks do not prove any of the higher concurrency targets.

## New message acceptance on October 17 2026

New message now opens in the main conversation area with a searchable To picker containing people and teams. People open or create a Direct Message; teams open the existing team conversation. The picker supports keyboard navigation, loading/error/retry states, cancellation, and paginated results. The back control uses the shared ChevronLeftIcon.

The composer accepts a draft before a recipient is chosen. Selecting a recipient merges it with any existing destination draft, preserving text, rich content, and attachments. Sending and uploading remain disabled until a recipient exists. Async navigation and send guards preserve drafts and ignore stale completions after cancellation or navigation.

The local UI fixture verified team search and keyboard selection, draft transfer, person selection, and the full pane layout. This is not another live API delivery test. All 58 Chat tests, targeted ESLint, and the full Projects TypeScript check passed. React Doctor's changed-file scan reported two unrelated warnings in document/story components; the untracked Chat directory is outside that scan, so its result is not Chat coverage.

![New message with people and teams](/Users/joseph/.codex/visualizations/2026/10/16/01a143d5-a48c-70b3-aa30-972ca8358107/chat-ui/chat-new-message.png)

## Shared viewer implementation checkpoint — October 17, 2026

The first delivery stage is implemented locally. Chat messages, replies, and Files & links now open the same controlled `AttachmentPreviewDialog`. The Chat provider lives above virtualized rows, resolves the source message for Files & links, preserves attachment order, excludes audio from modal navigation, and restores focus on close. Multiple files have finite previous/next controls and a counter; one file has neither. Removed files and failed authorization replace visible content with an error and retry where appropriate.

Message layouts now show one bounded image at its natural proportions, a compact responsive gallery with up to four columns and four mounted tiles, a rounded `+N` overflow tile, and compact wrapping document cards. Attachment corners use the smaller shared radius. One or two PDF attachments show the first page directly under their file headers, with the shared pink/red PDF vector icon. Larger and mixed document groups use compact headers. Images load lazily. Video opens in an active-only player. The later voice delivery now uses the shared inline audio player described in the implementation appendix below; audio stays outside modal attachment navigation.

Inline PDF covers are rendered client-side with the existing PDF.js assets only within 240 pixels of the viewport. A shared queue limits downloading/rendering to two jobs, renders only the top of page one at a bounded canvas size, cancels obsolete work, and destroys each PDF document/worker after painting. The small canvas remains while its row is mounted. Auto-fetch and streaming are disabled to avoid downloading unrelated pages when byte ranges are supported. This is an interim client preview, not a persisted server thumbnail; repeated row mounts may render again, and original PDF decoding can still exceed the final canvas memory budget.

Link cards now use the shared `components/ui/link-preview` renderer across Chat and story links. A separate `previewImage` field selects OG/Twitter imagery through the existing image proxy, while favicon-first `image` stays separate. Message URLs are derived once during API decoding when the server omits its link list, respecting explicit links and excluding code. Up to four unique HTTP(S) cards mount per message. A full preview appears only for one unique link with no other attachments; multiple links or any attachment use compact favicon/domain/title cards. The destination fragment is preserved on navigation. Failed/missing metadata or images leave a usable link. External cache behavior and authorized native cards are detailed in the appendix below.

PDFs render in the app with selectable text, page number/total, previous/next page, zoom, fit page/width, and download. The PDF library is lazy-loaded on opening; only the current document/page is mounted. Rendering, text extraction, and worker tasks are cancelled/released on changes and close. Canvas output is capped at 8 million pixels, 4096 pixels per edge, and DPR 2. Source-image decoding can exceed the output budget; oversized embedded images are not silently omitted. Gallery originals are fetched on demand without prefetching every file. Server-generated thumbnails and large-corpus performance validation remain later stages.

`/api/chat/[workspaceSlug]/conversations/[conversationId]/attachments/[attachmentId]/preview` provides a same-origin streaming PDF response. It authenticates the session, delegates conversation/file authorization to the existing Go attachment endpoint for every request, and uses its short-lived storage redirect without forwarding application cookies to storage. It validates byte ranges, PDF content type, response length, the existing 25 MiB limit, timeout, and cancellation. Responses are private/no-store. This makes PDF rendering independent of storage CORS and native-browser PDF plugins, at the cost of routing PDF bytes through the web server. Measure that bandwidth before a larger rollout.

The pinned `pdfjs-dist` release is 6.4.299. Projects dev/build scripts copy its matching legacy worker, CMaps, fonts, and WASM into ignored `public/pdfjs/6.4.299/`; production packaging must retain this public directory. `pnpm --filter projects prepare:pdf-assets` prepares an already-running dev environment. PDF.js requires Node 22.13 or newer; the repository CI Node 22.23.2 meets that requirement. Do not deploy mismatched library/worker versions.

Timeline initialization also waits for the first actual messages before starting Virtuoso's inverse positioning, while leaving the empty conversation/thread header usable. Existing cursor pagination, window retention, prepend anchoring, and sticky date behavior remain in place.

Validation includes 382 passing tests across 45 targeted Chat/viewer/metadata/API suites, scoped ESLint/formatting, and the full Projects TypeScript check. Browser checks covered an eight-image gallery, selected-image positioning, arrow navigation, zoom, Escape/focus restoration, single-PDF controls, three-page rendering, fit width, dark/light themes, and a compact viewport. The follow-up checks covered four-column galleries, two inline PDF covers, navigation between those PDFs, and the new-message footer. A generated three-page PDF was uploaded and sent as a local sample in Second → Team 1, then rendered through the real authenticated local API. A local message containing the public FortyOne URL also displayed its actual OG image through the metadata proxy. These verify local paths; they are not production or load-test evidence. React Doctor's changed-file scan reports two existing document/story warnings and does not cover the untracked Chat/viewer files, which have separate lint/tests.

Still outside the delivered viewer: server PDF first-page thumbnails, image derivatives, the full upload-draft redesign, Download all archives, Files & links list virtualization, and migration of the existing story viewer. The shared viewer is reusable without importing story ownership/deletion behavior. The viewer itself needs no database migration; the subsequent voice/transcription delivery uses migration 231.

![Shared PDF viewer through the local API](/Users/joseph/.codex/visualizations/2026/10/16/01a143d5-a48c-70b3-aa30-972ca8358107/chat-ui/06-local-api-pdf-viewer.png)

## Preview and voice implementation appendix — October 17, 2026

### Shared previews and typography

Chat, its composer, generated preview cards, and the shared attachment viewer use the application's body font and normal text sizing. Supporting copy no longer relies on the small `text-xs`/`text-sm` treatment. The single-link/no-attachments rule above controls full versus compact previews; full OG imagery is optional enrichment, and a favicon is never stretched into the social image.

External metadata starts near the viewport, with four concurrent browser requests and cancellation when the last observer leaves. Chat and story links share normalized query keys: fragments are removed for metadata lookup, while meaningful query strings are preserved. Successful data stays fresh for one hour; empty/error fallbacks become stale after 30 seconds. Inactive browser entries retain the existing 24-hour lifetime.

The server service uses a **per-process LRU of 1,000 entries**, at most **32 distinct in-flight URLs**, and one shared fetch for concurrent requests for the same normalized URL. Successful metadata has a one-hour TTL; failures and empty results have a 30-second TTL. Admission failures return retryable 503 responses and are not cached. One caller cancelling does not cancel other subscribers; the last cancellation aborts upstream work. Cached fields and URL lengths are bounded. This cache resets with its Next process and is **not shared across replicas**; Redis reuse has not been implemented.

Every route request authenticates before cache access. Responses are private/no-store; cache misses retain protocol/credential restrictions, public-address checks, pinned DNS, redirect checks, response-size limits, and timeouts. Images still use the authenticated raster-image proxy. Private native resources never enter this public metadata cache.

### Native FortyOne cards

Recognized story/task, objective, and document URLs use the corresponding authenticated resource APIs. The card module and its data load near the viewport; unsupported private app routes stay as safe FortyOne links rather than being sent to the external metadata fetcher. Access queries are scoped by current identity and target workspace and recheck access on mount before displaying cached details.

Cards subscribe to the application's canonical detail caches so existing optimistic updates remain visible. Full cards in the current workspace reuse title, status, priority, and assignee/lead controls for eligible team members; compact cards do not mount those field controls. Archived/deleted stories and users without the necessary role/membership do not get edit controls. Cross-workspace cards direct the user to open that workspace before making changes. Document cards display a bounded summary and open the document. Existing mutation endpoints remain authoritative for access and validation.

### Voice capture, playback, and transcription

Chat has a separate recorder lifecycle; Maya's dictation hook is unchanged. An explicit microphone click requests access after pausing shared playback. Recording stops at five minutes or rejects data exceeding **20 MiB**. MIME negotiation prefers Opus WebM and falls back to supported MP4; final recorder chunks are collected before review. The local review offers native play/pause, Discard, Record again, and Attach voice note. Stop and attach never send the message automatically. A failed attachment keeps the recording available for retry.

Sending is disabled while recording/review is open, and new-message recording requires a selected recipient. Upload completion merges with the latest text/rich-content/attachment draft; late results from another conversation are ignored and their pending uploads retired. Capture tracks and timers are released when recording finishes or is cancelled. Review object URLs are revoked on discard, re-record, successful attachment, and unmount. Recorded uploads use the existing attachment route with `kind=voice-note` and `durationMs`; sent metadata carries these optional fields. The current audio formats are audio-only WebM and MP4/M4A, not an unrestricted audio allowlist. Selected audio files do not acquire voice-note intent.

The server enforces the 20 MiB audio ceiling and validates supported containers. For recordings it checks declared duration and any duration available in the container. **Streaming WebM may omit a reliable duration**, so this is not a proof that a malicious client supplied at most five minutes of audio; the browser limit and server byte/container checks are the implemented boundary. Independent media probing/transcoding remains future work.

One shared audio element is created only after Play. Inline controls provide play/pause, seek, elapsed/total time, 1×/1.5×/2× speed, and download. Playback survives row virtualization and exposes a persistent return/stop bar; starting another note or local review coordinates playback. Switching workspace/account/conversation tears it down. Authorized source rechecks stop playback after deletion, attachment removal, or access loss; stale responses cannot stop a newly selected note.

**Transcription is explicit, after sending.** The first click checks for an existing saved result; only an unrequested or explicitly retried failed job issues POST. Same-tick and duplicate-row clicks are coalesced, with backend idempotency as the durable boundary. Pending/processing results poll every 2.5 seconds while observed, resume from cached state after remount, and stop polling on completion/error or unmount. Errors expose Retry instead of leaving a pending spinner. Ready text can expand/collapse without another inference request; empty ready text explains that no speech was detected. Automatic transcription, timed transcript seeking, transcript search/editing, and Maya consumption of voice transcripts remain outside this delivery. Waveforms were added in the subsequent refinement below.

### Evidence and remaining validation

- Web recording/upload/review lifecycle: **13 focused tests passed**, using mocked MediaRecorder, microphone, and transport behavior.
- Transcription/access/player lifecycle: **32 tests passed** at the integration checkpoint, including explicit requests, saved-result reuse, failed retry, cancellation, polling after remount, duplicate actions, source deletion/access loss, and playback teardown. The subsequent focused transcription rerun passed **10 tests** after correcting an outdated route-path assertion. The final combined selection passed **309 tests across 37 suites**, covering Chat, attachment/native previews, story links, and metadata routes. The latest audio lifecycle selection passed **33 tests**.
- Shared metadata/cache/security paths: **99 tests passed**; targeted ESLint and the full Projects TypeScript check passed at this checkpoint. Explicit-file ReactDoctor scans covered the new untracked voice sources and reported no errors, with the existing ChatShell complexity advisory; the recorder's object-URL advisory is covered by cleanup tests. The final explicit scan of new voice/native preview sources reported no errors and retained maintainability advisories. Broader scans still have baseline errors/advisories outside voice, and a tracked-changes-only scan does not cover all new untracked Chat files.
- Browser fixture recording → stop → native-audio review → attach was exercised with **fake MediaRecorder and synthetic AAC**. A separate synthetic AAC request successfully reached the real OpenAI mini transcription provider. Neither exercise captured the user's microphone, established real-device capture/playback compatibility, or completed a production acceptance test.

The real application at `http://localhost:3000/second/chat` separately displayed an authorized native card for **TM-267**. An ordinary generated M4A file then completed **upload → send → play → Transcribe** through the real API on port 8000, the worker on port 8080, and the real transcription provider. The visible result was: “This is a voice note preview test. The product review is tomorrow at 10.” This verifies the ordinary audio upload path without falsely marking the file as a microphone recording. The local worker was gracefully refreshed with the combined implementation; one healthy worker on port 8080 was confirmed.

![Generated M4A transcribed through the local app, API, worker, and provider](/Users/joseph/.codex/visualizations/2026/10/16/01a143d5-a48c-70b3-aa30-972ca8358107/chat-ui/11-live-transcript.png)

After a full-page reload, clicking Transcribe reopened the same saved transcript immediately. The approximately five-second sample also passed live play/pause. The UI result establishes persisted text retrieval; absence of another provider call is covered by the GET-ready/no-POST unit contract, not inferred from the screenshot. No actual user microphone was granted or captured. Real microphone permissions/capture, Safari and mobile recording/playback, long recordings, actual-storage seeking/expiry, deployment quotas, and sustained load remain acceptance work. These local synthetic-audio results are not production or real-device capture proof. The local app still reports an unrelated hydration mismatch in the shared notifications badge; this delivery does not modify that control.

### Slack research applied

Slack distinguishes classic crawled link previews from app-specific unfurls driven by `link_shared` and `chat.unfurl`. Its Work Objects add typed entities and a separate details pane, including richer interactions and authentication choices. These are distinct capabilities, not names for the same OG image card. [Slack unfurling documentation](https://docs.slack.dev/messaging/unfurling-links-in-messages/), [Work Objects overview](https://docs.slack.dev/messaging/work-objects-overview/)

FortyOne adopts that distinction through safe external metadata cards and authenticated native resource cards with existing edit controls. It does not implement Slack's event/API protocol, entity schema, related-conversation aggregation, embedded third-party apps, or a Work Objects details pane. The single-link OG rule, compact mixed-content layout, default typography, lazy loading, and cache limits are FortyOne product/implementation decisions.

### Waveform and preview refinement — October 17

Voice attachments now use a compact single-row player: play/pause, a seekable waveform, time at the right, a transcript toggle, speed, and a download menu. The filename is accessible without a visible heading. Native range semantics preserve keyboard and touch seeking. The composer microphone follows the mention button. All new visible text inherits the app font size.

Waveforms use 48 normalized bins from real audio, loaded only after a player stays visible for 200 ms. One conversation-scoped queue fetches/decodes at a time and caches at most 64 tiny peak arrays; no blobs or decoded PCM enter the cache. Canceled/offscreen requests are aborted; in-flight browser decoding cannot be canceled, so its semaphore remains held until completion and stale results are discarded. Downloads have a 15-second timeout and a streamed 20 MiB cap. Known durations over five minutes are skipped, and decoded results over five minutes are rejected. Browser decoding requires a complete file, so unknown or dishonest duration metadata is not a hard pre-decode memory limit; server-generated peaks remain the production-scale path for arbitrary long audio.

An explicit `?media=audio` mode on the existing authenticated attachment preview proxy supplies MP4/WebM waveform bytes without exposing storage URLs or forwarding application cookies to storage. The default route remains PDF-only. Audio playback and transcription are independent of waveform failures; unsupported files retain a neutral seek track.

Preview cards share a stable dark surface mixed from the background and surface tokens, slightly darker than a hovered message without becoming black. PDF thumbnail headers have no separator line. Story cards use the existing FortyOne logo, omit the redundant product name/title edit action, shorten raw internal links to references, and retain both whole-card and View story navigation. Both open the existing story editor in a slightly wider chat side panel with authorization checks, close/Escape handling, and focus restoration. View story uses primary-colored text without a button background. The panel reuses the story's own action header with a close button and a single content inset, avoiding a duplicate story-ID header. File icons use locally vendored Fluent and Office SVG assets with source/license notices in `components/ui/file-type-icons/NOTICE.md`.

## Proposed file previews and voice messages

Status: original design plan retained for follow-up work, October 17, 2026. The implementation checkpoints and appendix above take precedence: shared previews, native cards, recording, playback, and explicit transcription are now implemented locally. Media derivatives, expanded draft management, and other items not marked complete remain proposals. Voice required local migration 231; the current local schema also includes concurrent migration 232. The plan below is not evidence that every proposed behavior shipped.

The goal is to let people recognize, open, browse, listen to, and understand shared files without leaving Chat. Use FortyOne's typography, spacing, color tokens, rounded controls, and shared components. The supplied Slack screenshots establish the interaction model, including the difference between one document, grouped documents, image galleries, and inline audio.

### Original review before implementation

1. **Find shared files — useful foundation.** Files & links already has search, type filters, sorting, a media strip, and sender metadata. Preserve that structure. It has cursor pagination, but its loaded resource pages and rendered rows currently accumulate; include windowing and bounded retention in this work.
2. **Open a PDF — incomplete.** The current Chat dialog displays a file icon and an Open file link, rather than the document itself. Message attachment cards also open files externally. Replace these primary actions with the shared in-app viewer.
3. **Browse multiple attachments — missing in Chat.** A story attachment viewer already provides previous/next navigation, keyboard handling, and basic image/video/PDF rendering. Extract its presentation and controls without importing story-specific ownership/deletion behavior.
4. **Prepare attachments — functional but limited.** The composer uploads first and then shows filename chips. Add immediate visual drafts and independent upload states. Preserve typed text when asynchronous uploads finish.
5. **Record a voice message — not available.** Maya's recorder captures audio for dictation, but Chat does not accept or retain audio messages. Reuse its microphone lifecycle; add durable audio attachments and a player.

The screenshots below record the earlier UI-fixture baseline with sample data. They establish the visible file-list and PDF-dialog gaps. Native playback, microphone behavior, and large-history media performance require the acceptance tests below; screenshots alone do not establish those behaviors or accessibility compliance.

![Current Files and links](/Users/joseph/.codex/visualizations/2026/10/16/01a143d5-a48c-70b3-aa30-972ca8358107/chat-ui/01-current-files.png)

![Current PDF dialog shows an icon rather than the document](/Users/joseph/.codex/visualizations/2026/10/16/01a143d5-a48c-70b3-aa30-972ca8358107/chat-ui/02-current-pdf-preview.png)

### Attachment layout rules

These are initial dimensions for review, not new global typography tokens. All widths are maximums and collapse to the available message/thread width.

| Content in one message                           | Timeline presentation                                                                                                                                      | Open action                                                                           |
| ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| One PDF and no other attachments                 | File header with PDF mark, filename, type, size; bounded first-page image below. Start around 420px maximum width and 320px thumbnail height.              | In-app PDF viewer. No attachment navigation; PDF page/zoom controls remain.           |
| Multiple documents or PDF mixed with other files | Compact document cards, initially preserving the existing 220px cap. Wrap naturally; use one column in narrow threads. No tall PDF previews for each file. | In-app viewer with the message's eligible attachment sequence.                        |
| One image                                        | Preserve its aspect ratio, with proposed caps around 480px wide and 360px tall. Portrait images remain portrait; do not stretch a phone screenshot.        | Full image viewer, no attachment navigation if it is the only eligible attachment.    |
| Two to four images/videos                        | Bounded responsive gallery, stable tile dimensions, video play badge.                                                                                      | Open the selected item at its original position.                                      |
| More than four images/videos                     | Show four tiles. Final tile shows the hidden count: eight images means a +4 overlay.                                                                       | All eight remain available in the viewer.                                             |
| Audio file or recorded voice note                | Inline player with play/pause, time, seek control, speed, and optional transcript.                                                                         | Play inline; never open an image/document preview modal.                              |
| Unsupported document                             | Clear file-type card with filename/size.                                                                                                                   | In-app information/fallback view with explicit Download; do not pretend to render it. |

Mixed messages group visual media, documents, and audio without losing the original attachment order stored on the message. Attachment headings use a count and a shared chevron to collapse/expand the group. A Download all action can follow as a separate bounded server archive job; avoid launching ten browser downloads. It must recheck every attachment's access, sanitize filenames, handle duplicates, and expire temporary archives.

Use one shared file icon resolver across Chat, Files & links, composer previews, and comments. The requested recognizable PDF mark is a separate asset decision from card styling. The current asset is Microsoft's Fluent PDF file icon. Do not extract pixels or SVG from Slack. Source an appropriate PDF asset and preserve its provenance; Adobe distinguishes its PDF link icon from its Acrobat product logo and imposes specific reuse conditions. The exact Slack-style recolored mark cannot be assumed to be covered by Adobe's downloadable icon terms. [Adobe icon guidance](https://www.adobe.com/legal/permissions/icons-web-logos.html)

### In-app attachment viewer

Introduce one controlled AttachmentPreviewDialog above the virtualized rows. Its session contains the conversation ID, message ID, selected attachment ID, ordered eligible attachment IDs, and return-focus target. Rows request that it open; they do not own its lifetime.

Here, eligible means every non-audio attachment from that message, including unsupported documents shown with an honest fallback. Links extracted from the message are not attachments. The viewer counter counts this sequence, while the timeline's total attachment count includes audio. A mixed message with eight images and one PDF therefore has nine viewer items; adding an audio file does not create a tenth modal item. Label the counter as files when types are mixed.

- Scope file navigation to the attachments on the selected message. Do not accidentally browse unrelated messages or the entire channel. Opening from Files & links resolves the source message and follows the same rule.
- Eight images expose Previous, Next, keyboard arrows, and a 1 of 8 counter. Start on the clicked image. Use a consistent finite sequence with disabled controls at its ends.
- One eligible file hides the attachment counter and navigation buttons entirely. A single multi-page PDF still exposes its separate page navigation.
- Images use contain sizing, fit-to-window, zoom, and pan when zoomed. PDF viewing offers page number/total, next/previous page, zoom, fit width/page, and Download. Only the active PDF is mounted, and only its visible/nearby pages are rendered.
- Keep attachment navigation and PDF page navigation visually distinct. Label controls explicitly for assistive technology. Arrow-key file navigation must not steal keys from page inputs, text selection, or other focused controls.
- Audio attachments are excluded from the modal sequence and remain inline. Mixed image/PDF groups can navigate between those types; filenames and file-type labels make the transition clear.
- Escape closes the viewer. Restore focus to the original card when mounted; otherwise return to the message or timeline container after its virtual row has unmounted. Keep the user's scroll position.
- Deletion, access revocation, or a failed download shows an understandable state. It must not expose stale signed URLs or leave a broken blank modal.
- Do not prefetch all originals. Preload at most one adjacent image within a byte budget. Do not pre-render adjacent PDFs or video streams.

Extract the reusable behavior from StoryAttachmentPreview into a controlled shared component, keeping story deletion/ownership in its existing wrapper. Chat supplies its own authorized URLs and actions. Existing story/comment behavior needs regression checks after this extraction.

### Draft attachment UX

Create a draft upload controller independent of the editor and timeline. Each entry has a stable local ID, original File, preview URL, type, local dimensions, status, error, upload cancellation handle, and eventual attachment ID.

1. File selection or paste immediately creates a thumbnail or file card in the composer. Images use short-lived object URLs. PDF/document drafts stay compact until send; the first version generates their thumbnails only after the message is committed, rather than rendering PDFs in the editor or starting draft conversion jobs.
2. Upload with bounded concurrency, initially three files. A per-file spinner is honest with the current transport; show percentages only after adding actual upload-progress reporting.
3. A failed entry keeps its file and offers Retry or Remove. Successful files remain. Do not require reselecting the whole batch.
4. Send remains disabled while any retained file is unfinished or failed. State what needs attention. Removing a failed file restores eligibility.
5. Merge completion into the latest draft by local ID. The completed hook already uses the latest committed draft and operation generations to prevent stale uploads from overwriting later typing or attaching into another conversation. A full per-file upload controller remains follow-up work.
6. Abort removed/cancelled uploads. If the server still completes, retire the resulting pending attachment. Ignore results for a replaced draft or previous conversation.
7. Revoke object URLs when removed/replaced and when their owning draft is discarded. Revoke after playback/image consumers release them. Stop microphone tracks immediately on stop, cancel, or navigation.
8. Keep the current ten-file and 25MB-per-file general limits initially. New message attachment controls stay disabled until a person/team is chosen. Drafts must never migrate already-uploaded files into an unauthorized conversation.

### Voice recording and playback

Add the existing shared microphone icon beside the composer attachment/formatting actions. Label it Record voice message; this is distinct from Maya's existing Dictate text behavior.

Proposed recording flow: idle → waiting for microphone → recording → paused/recording → stopped for review → uploading → ready to send → sent. Stopping never sends automatically. The review state contains play/pause, elapsed/duration, a real level/waveform display if available, Discard, and Record again. Existing message text can accompany the audio.

The implemented Chat-specific hook handles permission races, final chunks, MIME negotiation, interruptions, and cleanup without changing Maya's dictation hook. The recording limit is five minutes and 20 MiB; see the appendix for the server duration caveat. Slack documents a five-minute clip limit; adopting it here is a product choice. [Slack audio and video clips](https://slack.com/help/articles/4406235165587-Record-audio-and-video-clips-in-Slack)

Microphone denial, no device, interrupted recording, and an unanswered permission prompt need distinct messages and a working Cancel action. Use a secure context, negotiate with MediaRecorder.isTypeSupported, retain the actual MIME type/extension, and wait for final dataavailable/stop events. Chunk delivery is not a reliable elapsed-time clock; calculate time independently and verify duration during server probing. [MDN microphone access](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia), [codec detection](https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder/isTypeSupported_static), [recording chunk behavior](https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder/dataavailable_event)

The sent player uses a shared controller above the virtualized list. Only one recording plays at once; starting another pauses the previous one. It uses play/pause, keyboard-accessible seek, elapsed/total time, 1x/1.5x/2x speed, and a transcript toggle. No autoplay. While playback's row is offscreen, show a compact persistent Now playing control with Pause and Return to message; switching the workspace/account/conversation stops playback and releases its source. Retaining position across those navigation boundaries is not implemented. Deleting the source stops playback. Audio never opens the document modal.

Generate real waveform peaks once in the media worker and store a small normalized array, initially 64–128 bins. Do not decode every audio file in every client's browser. Until peaks are ready, use a simple seek bar rather than invented waveform data. Uploaded audio uses the same player even when it is not marked as a recorded voice note.

### Transcription and Maya

Run transcription after a voice note has been sent and accepted. Audio remains independently playable; a provider outage must not turn a successful voice message into a failed send. The transcript starts collapsed and exposes queued, processing, ready, failed/retry, disabled, or no-speech states. Label generated text as a transcript and keep the original audio as the source of truth.

The existing /api/transcribe route is synchronous dictation: it accepts up to 5MB, uses gpt-4o-mini-transcribe, and returns text. Reuse provider configuration and billing patterns, but process durable voice messages by authorized attachment ID in a worker. Do not point that endpoint at arbitrary file URLs or make the client repeat transcription on every transcript expansion.

The current official guide recommends gpt-transcribe for ordinary recorded audio. Compare it with the already-integrated mini model using names, numbers, local accents, mixed languages, background noise, silence, and short recordings before choosing the default. Persist the model and processor version. Keep provider selection behind a small service interface. [OpenAI speech to text](https://developers.openai.com/api/docs/guides/speech-to-text)

Plain text is the first release. Do not invent timestamps from word count or show synchronized highlighting without timing data. The official capability guide directs word/segment timestamps to whisper-1 and speaker-labeled segments to a diarization model; choose a timing-capable pipeline if click-to-seek becomes a requirement. [OpenAI transcription capabilities](https://developers.openai.com/api/docs/guides/transcription#choose-a-specialized-capability)

Superseded default: the original plan proposed automatic transcription, but the implemented product requires an explicit Transcribe action after sending. Any future automatic mode needs a separate workspace setting and per-recording choice before send. Discarded drafts are never sent to the transcription provider. Apply per-user/workspace minute budgets and concurrency limits; exhausting the budget must still allow sending/listening to audio. Do not log audio, transcript text, or signed URLs.

Maya may use an available transcript only when explicitly invoked with the existing conversation permissions. Spoken words such as “at Maya” must not create a structured mention or trigger an action. If the requested note is still transcribing, display a specific pending state and resume a bounded invocation when ready; avoid guessing its content. A later enhancement can let Maya summarize a voice discussion or propose tasks with links to the source message, using existing approval and task-rendering flows. Transcript indexing and edits should be separate follow-up work with deletion propagation and the same access filters as message search.

### Data and processing architecture

Keep the existing attachment identity, pending-owner association, transactional message claim, object storage, and retention outbox. For new uploads preserve the original bytes. The current JPEG/PNG optimizer overwrites its source; do not reuse that behavior for previews. Historical originals already overwritten cannot be reconstructed by a migration.

Proposed additive schema, with migration names/numbers assigned only at implementation time:

| Entity                    | Data and invariants                                                                                                                                                                                            |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Attachment media metadata | attachment/workspace IDs, source version/hash, probed MIME/container/codec, width/height, duration, page count, voice-note intent, processor version. Server values override untrusted client metadata.        |
| Attachment variants       | attachment ID, source version, kind, processor version, private object key, dimensions, MIME, bytes, status, safe failure code. Unique key prevents duplicate derivatives.                                     |
| Media processing jobs     | work kind, durable state, attempts, retry time, lease token/expiry, source version, timestamps. Atomic lease/commit avoids multiple workers publishing incompatible results.                                   |
| Attachment transcripts    | attachment/workspace IDs, source version, model/processor version, status, language, bounded text, optional supported timed segments, completion time, metered minutes. One active result per source/pipeline. |

Use new migrations; do not alter applied migrations 229–232. Update SQLC, domain types, API serializers, client decoders, cache updates, and retention logic together. Existing attachments receive nullable metadata and keep working as compact cards. Avoid downloading or processing the entire historical library during deployment. Generate missing previews lazily or with a throttled resumable backfill, prioritizing recently active conversations.

Processing sequence:

1. Authorized upload writes the original and pending attachment association. Validate file size, bytes, MIME/container, and supported format; do not trust filename/MIME claims alone.
2. Send transaction claims attachments and writes processing work durably. Enqueue after commit; a recovery sweep finds committed but unenqueued jobs.
3. A media worker probes and creates supported derivatives: bounded image thumbnails, PDF page-one image, video poster, audio duration/peaks, and a browser-compatible playback variant if required.
4. A separate transcription worker consumes only eligible sent recordings. It stores results and actual usage. Retries use finite backoff; a timed-out provider request may still be billable, so internal idempotency is not a promise of exactly-once provider billing.
5. Commit derivative status and a recipient-scoped update event together. Clients refresh authorized metadata; viewers and players update without replacing the message or changing its timestamp.

Use separate bounded media and transcription queues/processes. Current workers share concurrency ten with other jobs; CPU-heavy PDF/media conversion must not block notification delivery or Maya. Start conservatively, for example two conversion processes per worker deployment and a separately configurable transcription concurrency, then tune from memory/CPU and provider measurements.

For PDFs, render page one once as a cached image in an isolated renderer; lazy-load PDF.js only for an opened full viewer. PDF.js supports page retrieval and bounded canvas rendering. Select/pin the renderer and runtime after measuring representative files, and disable document scripting and external resource execution. Cap decoded pixels, pages inspected, CPU, memory, temporary disk, and runtime. Encrypted/corrupt/unsupported PDFs fall back to a usable card and an explicit explanation. [Mozilla PDF.js examples](https://mozilla.github.io/pdf.js/examples/)

Image, PDF, and audio support come first. Plain text/CSV can use bounded escaped previews. Office documents need a separately isolated conversion step before claiming in-app Word/Excel/PowerPoint previews; until then show an honest download fallback. Do not send private documents to a public remote Office/PDF viewer.

The current direct-upload path allows unscanned attachments. Before running converters on arbitrary uploads, establish the intended quarantine/scanning policy and an isolated processing boundary. Originals, variants, and transcripts must all respect the same eligibility/access state. Confirm what scanning actually runs in deployment rather than relying on an existing comment or status column.

### API and real-time changes

Retain current HTTP uploads and message writes. No WebSocket service is needed. Add metadata to existing attachment responses and use the existing scoped SSE connection for completion hints.

| Proposed operation                   | Behavior                                                                                                                                                                                                   |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GET conversation attachment metadata | Returns media properties, variants, processing states, and transcript availability after current membership checks. Bulk metadata should ride in message/resource responses to avoid one request per card. |
| GET attachment variant               | Authorizes the original relationship, then serves/redirects to a private derived object. Never accept arbitrary storage keys or a remote URL.                                                              |
| GET attachment transcript            | Returns authorized persisted text/status; lazy-load text when expanded rather than embedding every transcript in the timeline payload.                                                                     |
| POST transcript retry                | Rechecks entitlement, ownership/policy, source state, and quota. Reuses the existing result/job when applicable; cannot create duplicate messages.                                                         |
| GET original with download mode      | Explicit original download disposition and safe filename; do not rely solely on a browser download attribute across origins.                                                                               |
| SSE attachment update                | Small attachment/message identifiers plus monotonic asset version. No raw file bytes, transcript body, or storage key.                                                                                     |

Use the current durable outbox and recipient-scoped delivery. Preview/transcript completion does not generate another chat unread count, email, push notification, or Maya invocation. Coalesce multiple updates for the same message and refresh only affected metadata/cache entries. Missing SSE events are repaired by normal refetch and low-frequency polling only while an open item is processing; stop on completion, hidden view, access loss, or unmount.

Pending uploads have no committed message ID, while today's chat outbox requires one. Show local draft status and use owner-only status polling where needed before send; do not pretend those states already fit the message outbox.

Current authorized downloads redirect to a one-minute signed URL. Verify range requests, 206 responses, CORS, MIME, and content disposition in the actual storage provider. Use an authorized renewal path that can resume at currentTime after expiry. If provider redirects cannot support reliable seeking/refresh, add a bounded streaming proxy with range support; never buffer an entire recording into API memory. A signed URL remains usable for its lifetime, so immediate revocation requires a stronger proxy/token scheme. Avoid putting private media into a public image-optimization cache.

### Performance plan

Message history already uses keyset/cursor pagination. Server history queries compare sequence boundaries, order by sequence, and limit results; they do not walk increasing OFFSET values. The frontend uses before/after/around cursors, a normal page size of 50, and eight retained pages per message query. Its Virtuoso/GroupedVirtuoso timeline renders a viewport window with date groups and loads older history near the top.

That is approximately 400 retained messages per feed query, not a global browser-memory limit. The separate latest-50 query can add overlapping tail data, and replies, search anchors, and conversations have their own caches. The shared QueryClient currently retains inactive queries for 24 hours. Add shorter Chat-specific retention and a bounded working-set policy, while preserving drafts and intentional navigation history. Cache eviction must not discard unsent drafts or active viewer/playback state.

Required extensions:

- Keep stable message/attachment keys and reserve image/PDF aspect ratios before assets arrive. Test scroll anchoring when old pages prepend or evict and when thumbnails/transcripts expand.
- Render lightweight thumbnails in rows. No embedded PDF engines, full-resolution image downloads, waveform decoding, or active audio elements for every historical attachment.
- Own the document viewer and audio playback outside the virtual list. Scrolling a row away must not close the modal or stop playback unexpectedly.
- Update Files & links with a virtualized list/grid and bounded page window. Maintain a visible anchor during eviction; adding maxPages alone would silently remove visible earlier rows. Give the conversation picker/list an intentional large-directory policy too.
- Lazy-load the PDF viewer/transcription panel. Load only the active PDF and nearby visible pages; cap canvas dimensions and dispose page canvases, render tasks, object URLs, and workers on close.
- Use separate small and medium image variants sized for cards and galleries. Keep originals for explicit viewing/download. Generated variants should be cached privately and keyed by immutable source/pipeline versions.
- Bound uploads, prefetches, retries, renderer processes, and provider requests. Cancel obsolete browser requests. Avoid one polling loop or SSE connection per attachment; reuse the workspace connection.
- Coalesce status events and patch affected entities. A thumbnail becoming ready must not refetch every conversation or all historical pages.
- The existing SSE handler coalesces updates but broadly invalidates conversation queries. Add attachment-aware refresh and sustained backpressure before sending frequent processing updates. Validate resource-search query plans separately from the already-tested message-history plans.
- Measure p50/p95 query latency, client memory, mounted row count, long tasks, scroll stability, bytes fetched, queue age, processing failures, storage/egress, and transcription minutes.

Initial capacity planning remains 1,000 concurrent users. Validate 5,000, 10,000, and 50,000 as separate load stages rather than extrapolating from connection counts. Define room size, active senders, uploads/minute, file-size mix, voice minutes/minute, simultaneous playback, and geographic latency for each stage. A burst of large PDFs can require more worker capacity than many idle SSE users. Gate each stage on measured results and scale media workers independently of API/SSE replicas.

### Acceptance tests and release gates

Proposed performance budgets must be measured on named devices, browsers, network profiles, and representative files; they are targets, not current guarantees.

| Test                 | Acceptance                                                                                                                                                                                                                                           |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| One PDF              | First-page thumbnail or honest pending/fallback card; in-app readable PDF; page controls present; attachment navigation absent.                                                                                                                      |
| Eight images         | Four visible tiles with +4; every image reachable in original order; starts on the selected item; no unrelated message attachments.                                                                                                                  |
| Mixed files          | Bounded layout; image/PDF navigation correct; audio stays inline; count and keyboard labels are unambiguous.                                                                                                                                         |
| File dialog lifetime | Remains open while its source row is virtualized away. Closing restores a valid focus target and preserves timeline position.                                                                                                                        |
| Audio lifetime       | One active playback; scrolling keeps controls available; seeking/speed work; a single audio attachment never opens a modal.                                                                                                                          |
| Upload races         | Typing survives completion; removal/cancel does not resurrect files; partial failure can retry individually; switching conversations cannot misattach a file.                                                                                        |
| Recording            | Permission denial, delayed permission, no mic, device removal, stop/cancel, five-minute boundary, Safari MP4, Chromium WebM, and background interruption preserve understandable state.                                                              |
| Transcription        | Failure leaves playable audio; expand never rebills; retries are bounded; no-speech does not invent a message; no automatic Maya mention.                                                                                                            |
| Access and deletion  | Cross-tenant, wrong conversation, revoked membership, pending-owner access, and deletion during processing cannot expose originals/variants/transcripts. Workers cannot resurrect deleted files.                                                     |
| Long history         | Seed at least 100,000 historical messages with a realistic mixed-media distribution. Mounted row count and retained active-feed pages stay bounded; no full-history query/download.                                                                  |
| Scroll and memory    | No visible jump when paging or thumbnails arrive; no sustained memory growth after repeated open/close/conversation switches; profile 30-minute scroll/playback sessions.                                                                            |
| Responsiveness       | Target immediate local draft feedback within 100ms; ordinary input/scroll must not be blocked by PDF processing. Investigate repeated main-thread tasks over 50ms on the reference desktop.                                                          |
| Service latency      | Under the agreed workload, target p95 history/metadata reads below 300ms, send acknowledgement below 500ms excluding upload, and steady-state SSE visibility below one second. These are load-test gates, not proven capacity.                       |
| Async turnaround     | Establish measured p95 preview/transcript budgets by file size/duration, initially aiming for a small PDF thumbnail within 10s and a one-minute note transcript within 30s under launch load. Queues expose pending state when budgets are exceeded. |
| Accessibility        | Keyboard-only picker/gallery/player, visible focus, named controls, seek semantics, transcript readability, reduced motion, narrow threads, 200% zoom, light/dark themes. Do not announce playback time every second to screen readers.              |
| Delivery             | SSE reconnect/replay and missed-event refetch converge without duplicate messages, unread changes, notifications, or processing work.                                                                                                                |

Use local sample PDFs/images/audio and deterministic provider fakes for automated tests. Benchmark real transcription separately with representative consented samples; browser fixture checks do not prove storage, permission, provider, or production performance behavior.

### Delivery sequence

1. **Shared UI contracts and viewer.** Extract file identity/viewer primitives, implement the single-versus-group rules with fixtures, in-app image/PDF navigation, focus handling, and audio player states. Review light/dark and narrow threads before wiring more processing.
2. **Safe media foundation.** Add additive metadata/variant/job migrations, immutable originals for new files, authorized variant access, processing isolation, retention, and worker recovery. Deploy API/read compatibility before workers/backfill.
3. **PDF/image previews and upload drafts.** Generate thumbnails, wire per-file progress/retry/cancel, integrate Files & links and message groups, then validate virtualization and caches.
4. **Voice capture and delivery.** Allow validated audio, configure recording limits, persist voice-note intent, deliver playable audio, waveforms/seek/speed, expiry handling, and the shared playback controller. Ship behind a workspace flag.
5. **Transcription.** Add durable results, usage limits, retry UI, workspace settings, SSE refresh, and quality evaluation. Enable independently of voice notes.
6. **Load and resilience gates.** Exercise real storage, permission boundaries, worker restarts, provider failures, long histories, and the 1,000-user workload. Increase the enabled cohort only when latency, memory, failure, and spend targets hold.

Keep preview, voice recording, and automatic transcription flags separate. Rolling back a new worker or UI flag must leave existing messages and original downloads usable. Deletion removes all derivatives/transcripts through durable cleanup; shared attachments remain until their last authorized reference is retired.

### Cost model

Budget storage for originals plus derivatives, egress for previews/playback/downloads, CPU for conversion/transcoding, database/queue work, and transcription separately. Generate derivatives once per file version and reuse one persisted transcript. Opening a preview or transcript must not trigger another provider call.

At the documentation prices checked for this plan, gpt-transcribe is listed at approximately $0.0045/minute and the existing gpt-4o-mini-transcribe at $0.003/minute. Illustratively, 10,000 transcribed minutes cost about $45 or $30 respectively; 100,000 minutes cost about $450 or $300, before retries, storage, egress, processing, and applicable provider charges. Recheck prices and actual usage before rollout. [OpenAI pricing](https://developers.openai.com/api/docs/pricing)

Measure demand as audio minutes and media bytes, not only concurrent users. A workspace budget, finite retries, queue backpressure, and useful failure states keep busy periods from turning into uncontrolled spend.
