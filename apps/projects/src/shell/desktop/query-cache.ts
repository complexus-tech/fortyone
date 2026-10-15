import type { QueryClient, QueryKey } from "@tanstack/react-query";
import type {
  DesktopCacheEnvelope,
  DesktopCacheScope,
  DesktopCachedQuery,
} from "./cache-types";

const MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;
const MAX_QUERIES = 250;
const MAX_BYTES = 8 * 1024 * 1024;
const WORKSPACE_FAMILIES = new Set([
  "stories",
  "documents",
  "teams",
  "members",
  "statuses",
  "labels",
  "sprints",
  "objectives",
  "notifications",
  "custom-fields",
]);
const isRecord = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === "object" && !Array.isArray(value);

export const isDesktopCacheKey = (key: QueryKey, scope: DesktopCacheScope) => {
  if (key[0] === "users")
    return (
      (key.length === 2 && key[1] === "profile") ||
      (key.length === 3 &&
        key[1] === "automation-preferences" &&
        key[2] === scope.workspaceSlug)
    );
  if (key[0] === "workspaces")
    return (
      (key.length === 2 && key[1] === "list") ||
      (key.length === 3 &&
        key[1] === "settings" &&
        key[2] === scope.workspaceSlug)
    );
  return (
    typeof key[0] === "string" &&
    WORKSPACE_FAMILIES.has(key[0]) &&
    key[1] === scope.workspaceSlug
  );
};

// Temporary upload previews belong to the current webview, never a durable cache.
const withoutTemporaryAttachments = (value: unknown): unknown => {
  if (Array.isArray(value))
    return value
      .filter(
        (item) =>
          !(
            isRecord(item) &&
            typeof item.id === "string" &&
            item.id.startsWith("temp-") &&
            ("url" in item || "filename" in item)
          ),
      )
      .map(withoutTemporaryAttachments);
  if (!isRecord(value)) return value;
  return Object.fromEntries(
    Object.entries(value).map(([key, child]) => [
      key,
      withoutTemporaryAttachments(child),
    ]),
  );
};

export const createDesktopCache = (
  client: QueryClient,
  scope: DesktopCacheScope,
  now = Date.now(),
): DesktopCacheEnvelope => {
  const queries: DesktopCachedQuery[] = [];
  let bytes = 0;
  const candidates = client
    .getQueryCache()
    .getAll()
    .filter(
      (query) =>
        isDesktopCacheKey(query.queryKey, scope) &&
        query.state.data !== undefined &&
        query.state.dataUpdatedAt > 0,
    )
    .sort((a, b) => b.state.dataUpdatedAt - a.state.dataUpdatedAt);
  for (const query of candidates) {
    if (queries.length >= MAX_QUERIES) break;
    if (
      query.queryKey[0] === "users" &&
      query.queryKey[1] === "profile" &&
      (!isRecord(query.state.data) || query.state.data.id !== scope.accountId)
    )
      continue;
    const entry = {
      queryKey: query.queryKey,
      data: withoutTemporaryAttachments(query.state.data),
      updatedAt: query.state.dataUpdatedAt,
    };
    let size: number;
    try {
      const serialized = JSON.stringify(entry);
      if (serialized.includes("blob:")) continue;
      size = new TextEncoder().encode(serialized).byteLength;
    } catch {
      continue;
    }
    if (size > MAX_BYTES - bytes) continue;
    bytes += size;
    queries.push(entry);
  }
  return { ...scope, version: 1, savedAt: now, queries };
};

export const restoreDesktopCache = (
  client: QueryClient,
  scope: DesktopCacheScope,
  value: unknown,
  now = Date.now(),
) => {
  if (
    !isRecord(value) ||
    value.version !== 1 ||
    value.accountId !== scope.accountId ||
    value.workspaceId !== scope.workspaceId ||
    value.workspaceSlug !== scope.workspaceSlug ||
    typeof value.savedAt !== "number" ||
    value.savedAt > now + 60_000 ||
    now - value.savedAt > MAX_AGE_MS ||
    !Array.isArray(value.queries) ||
    value.queries.length > MAX_QUERIES
  )
    return 0;
  let restored = 0;
  for (const candidate of value.queries) {
    if (
      !isRecord(candidate) ||
      !Array.isArray(candidate.queryKey) ||
      !isDesktopCacheKey(candidate.queryKey, scope) ||
      candidate.data === undefined ||
      typeof candidate.updatedAt !== "number" ||
      !Number.isFinite(candidate.updatedAt) ||
      candidate.updatedAt > now + 60_000 ||
      now - candidate.updatedAt > MAX_AGE_MS
    )
      continue;
    if (
      candidate.queryKey[0] === "users" &&
      candidate.queryKey[1] === "profile" &&
      (!isRecord(candidate.data) || candidate.data.id !== scope.accountId)
    )
      continue;
    if (JSON.stringify(candidate.data).includes("blob:")) continue;
    const current = client.getQueryState(candidate.queryKey);
    if (current && current.dataUpdatedAt >= candidate.updatedAt) continue;
    client.setQueryData(candidate.queryKey, candidate.data, {
      updatedAt: candidate.updatedAt,
    });
    restored++;
  }
  return restored;
};
