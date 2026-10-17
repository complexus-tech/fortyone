# Chat upload validation

Chat attachments and voice recordings use the existing private attachment
pipeline. They do not require ClamAV, another service, or a scanner environment
variable. `APP_CHAT_SCANNER_ADDRESS` has been removed from application configuration.

## Availability and authorization

The workspace's Chat feature setting controls Chat capabilities. Enabled Chat
workspaces can attach files and record voice notes. Authentication, active workspace
membership, conversation membership, and write permissions are still required for
each upload. Reads and transcription recheck access independently.

New uploads retain the truthful `unscanned` state and no scan-completion timestamp.
No malware scanning runs in this pipeline. Chat accepts the existing attachment
states `unscanned` and `clean`; `pending`, `infected`, and `failed` remain blocked.
Existing states and scan history are preserved without a backfill or migration.

## Request and storage behavior

- Keep the existing file-extension, detected content-type, empty-file, and size
  validation. Generic attachments are limited to 25 MiB. Audio uploads are limited
  to 20 MiB, and recorded voice notes to five minutes.
- Recorded audio must be a supported audio-only WebM or MP4 container. Container
  validation remains separate from the generic file upload allowlist.
- Persist original bytes in private object storage and record workspace/owner
  metadata. Unlinked uploads retain the existing orphan-cleanup lifecycle.
- Require workspace and conversation authorization before linking an upload to a
  message, reading the original or thumbnail, and requesting transcription.
- Serve originals through short-lived authorized access URLs. Generate image
  optimizations and private thumbnails through the existing worker lifecycle.
- Keep transcription request, retry, playback, and access-revocation behavior.
  Transcription uses the existing model configuration and needs no scanner.

## Validation

Service tests exercise attachment and recorded-audio uploads without any scanner,
including `unscanned` persistence, file validation, private storage, and cleanup.
Real PostgreSQL tests cover linking, message claiming, authorized reads, and
transcription for both allowed states; blocked states, deleted messages, revoked
membership, cross-workspace access, and ownership failures remain denied.

The earlier ClamAV deployment instructions and scan-test results are historical.
No ClamAV process or scanner environment variable is required for deployment.
