# Workspace Path-Based Routing Plan

## Goals

- Move the Projects app to path-based workspaces at `/{slug}/...`.
- Remove host/subdomain parsing across client and server code.
- Centralize workspace slug resolution with helpers.

## Plan

1. **Routing structure**

   - Move workspace-scoped routes under `app/[workspace]/...` so paths become `/{slug}/my-work`, `/{slug}/settings`, etc.
   - Keep root routes (e.g., `/login`, `/logout`, `/onboarding`) outside the workspace scope.

2. **Workspace slug helpers**

   - Add `useWorkspaceSlug()` (client) via `useParams()` for the first segment.
   - Add `getWorkspaceSlugFromParams(params)` (server) for layouts/pages.
   - Use helpers everywhere slug access is needed.

3. **Fetch wrapper update**

   - Update `apps/projects/src/lib/http/fetch.ts` to use path-based slug.
   - Client: derive from `useParams()` (or pathname fallback).
   - Server: accept slug via argument or helper provided by layout/page params.

4. **Workspace resolution**

   - Replace host-based lookup in `app/(workspace)/layout.tsx`, analytics, settings, and `getCurrentWorkspace`.
   - Match workspace by `params.workspace`.

5. **Navigation & redirects**

   - Update workspace switcher and invitation redirect URLs to `/${slug}/my-work`.
   - Update displayed workspace URL to `/${slug}`.
   - Introduce a helper like `workspacePath("/my-work")` for slug-prefixed links.

6. **Pathname checks**

   - Normalize pathname comparisons by stripping the slug prefix.
   - Replace strict checks like `pathname === "/my-work"` with normalized comparisons.

7. **Audit**
   - Re-scan for remaining `host`/subdomain usage and root-path assumptions.

## Open Questions

- Should `/` redirect to a default workspace or landing page?
- Best approach for server-side slug: passed through arguments vs shared context?
