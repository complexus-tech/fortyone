# Attachment authorization and storage safety

This document is the security review checklist for attachment operations. It complements `docs/database/attachments.md`, which explains the schema and SQLC adapter.

## Trust boundaries

There are four separate identities in an attachment workflow:

1. the authenticated principal;
2. the current workspace from trusted request context;
3. the owning story or document;
4. the storage object referenced by database metadata.

Code must never replace one identity with another. In particular, an attachment UUID, a blob name, a signed object URL, or a user-supplied workspace ID is not sufficient proof of access.

## Authorization matrix

| Operation                        | Guest                                 | Workspace member                              | Attachment uploader                      | Workspace admin                | Background worker                                        |
| -------------------------------- | ------------------------------------- | --------------------------------------------- | ---------------------------------------- | ------------------------------ | -------------------------------------------------------- |
| List story attachments           | route policy only                     | allowed for an authorized story               | allowed                                  | allowed                        | not exposed                                              |
| Resolve document/story media     | denied without owning-resource access | allowed after document/story policy           | allowed after resource policy            | allowed after resource policy  | not exposed                                              |
| Upload ordinary story attachment | denied                                | allowed by story route policy                 | allowed                                  | allowed                        | not exposed                                              |
| Delete ordinary story attachment | denied                                | denied unless uploader                        | allowed for the exact workspace relation | allowed in the exact workspace | not exposed                                              |
| Link inline story media          | denied                                | allowed only as a current workspace member    | allowed                                  | allowed                        | not exposed                                              |
| Delete orphaned media            | no direct route                       | only after owning module removes its relation | same                                     | same                           | not exposed                                              |
| Claim optimization               | not exposed                           | not exposed                                   | not exposed                              | not exposed                    | allowed with `(workspace_id, attachment_id)` and a lease |

Route and owning-resource policies remain responsible for deciding whether a member may edit a story or document. The attachment repository provides the second line of defense: every mutation proves the resource and attachment belong to the same workspace.

## Upload controls

- Multipart bodies are bounded before parsing.
- File size and filename length are bounded.
- The service detects content type from bytes; it does not trust only the multipart header or filename extension.
- Inline document/story media is restricted to images and MP4.
- Signed object access is short-lived and generated only after database authorization.
- Known infected, pending-scan, and failed-scan objects are not signed.
- Optimization work is bounded by image dimensions as well as bytes to limit decompression bombs.

## Remote profile images

Provider-supplied profile image URLs use `internal/platform/safehttp.Downloader` rather than `http.DefaultClient`.

The downloader requires HTTPS on port 443, rejects credentials, fragments, IP-literal and single-label hosts, rejects a DNS answer containing any private/special-use address, pins the connection to a validated address, verifies TLS against the original hostname, disables redirects, proxies, and connection reuse, and bounds time, headers, and response bytes. The image is still accepted only after byte-based type detection.

Do not weaken this policy to make a provider redirect work. Add a provider-specific, signed, allowlisted flow only after a documented threat-model review.

## Negative tests required for changes

- attachment ID from workspace B under workspace A context;
- story from workspace B with an attachment from workspace A;
- non-member `created_by` for inline story media;
- delete using the wrong workspace;
- resolve or sign an infected/pending/failed-scan object;
- duplicate and concurrent optimization claims;
- stale optimization completion;
- private, loopback, link-local, metadata, IP-literal, redirecting, oversized, and non-image remote profile targets;
- unlink while another supported relation still exists.

Security-sensitive failures should return not-found for mismatched tenant/resource identities, forbidden for an authenticated principal that lacks an allowed role on a known resource, and bad-request for malformed identifiers. Logs may include opaque attachment/workspace IDs and a request ID; they must not include signed URLs, storage credentials, authorization headers, or downloaded bytes.
