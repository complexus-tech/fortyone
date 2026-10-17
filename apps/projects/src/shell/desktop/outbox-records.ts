import type { QueryClient } from "@tanstack/react-query";

export const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value && typeof value === "object" && !Array.isArray(value));

export function setChanged(
  client: QueryClient,
  key: readonly unknown[],
  value: unknown,
) {
  if (JSON.stringify(client.getQueryData(key)) !== JSON.stringify(value))
    client.setQueryData(key, value);
}

const collections = new Set([
  "pages",
  "groups",
  "stories",
  "subStories",
  "comments",
  "subComments",
]);
export function findResource(
  value: unknown,
  id: string,
): Record<string, unknown> | null {
  if (Array.isArray(value)) {
    for (const item of value) {
      const found = findResource(item, id);
      if (found) return found;
    }
  } else if (isRecord(value)) {
    if (value.id === id) return value;
    for (const key of collections) {
      const found = findResource(value[key], id);
      if (found) return found;
    }
  }
  return null;
}

/** Identity remapping never touches text or arbitrary string values. */
export function remapResource(
  value: unknown,
  from: string,
  to: string,
  options: { preferCanonical?: boolean } = {},
): unknown {
  if (from === to) return value;
  if (Array.isArray(value)) {
    const result: unknown[] = [];
    let target: { index: number; canonical: boolean } | undefined;
    for (const original of value) {
      const item = remapResource(original, from, to, options);
      if (!isRecord(item) || item.id !== to) {
        result.push(item);
        continue;
      }
      const canonical = isRecord(original) && original.id === to;
      if (!target) {
        target = { index: result.length, canonical };
        result.push(item);
        continue;
      }
      const previous = result[target.index] as Record<string, unknown>;
      const newer =
        Date.parse(String(item.updatedAt)) >=
        Date.parse(String(previous.updatedAt));
      // A fetched comment owns its current thread, including an empty thread
      // after deletion. Never union a provisional comment's stale replies.
      const useIncoming =
        options.preferCanonical && canonical !== target.canonical
          ? canonical
          : newer;
      result[target.index] = useIncoming
        ? { ...previous, ...item }
        : { ...item, ...previous };
      target.canonical ||= canonical;
    }
    return result;
  }
  if (!isRecord(value)) return value;
  const result = Object.fromEntries(
    Object.entries(value).map(([key, item]) => [
      key,
      ["id", "storyId", "parentId", "resourceId"].includes(key) && item === from
        ? to
        : remapResource(item, from, to, options),
    ]),
  );
  if (Array.isArray(value.stories) && Array.isArray(result.stories)) {
    const removed = value.stories.length - result.stories.length;
    if (typeof value.totalCount === "number")
      result.totalCount = Math.max(0, value.totalCount - removed);
    if (typeof value.loadedCount === "number")
      result.loadedCount = Math.max(0, value.loadedCount - removed);
  }
  return result;
}

export function patchResource(
  value: unknown,
  id: string,
  fields: Record<string, unknown>,
  onlyIfNotNewer = false,
): unknown {
  if (Array.isArray(value))
    return value.map((item) => patchResource(item, id, fields, onlyIfNotNewer));
  if (!isRecord(value)) return value;
  let patched = value;
  if (value.id === id) {
    const currentTime =
      typeof value.updatedAt === "string" ? Date.parse(value.updatedAt) : NaN;
    const incomingTime =
      typeof fields.updatedAt === "string" ? Date.parse(fields.updatedAt) : NaN;
    const newer =
      onlyIfNotNewer &&
      Number.isFinite(currentTime) &&
      Number.isFinite(incomingTime) &&
      currentTime > incomingTime;
    patched = newer
      ? { ...value, desktopPending: false, desktopCreated: false }
      : { ...value, ...fields };
  }
  return Object.fromEntries(
    Object.entries(patched).map(([key, item]) => [
      key,
      collections.has(key)
        ? patchResource(item, id, fields, onlyIfNotNewer)
        : item,
    ]),
  );
}
