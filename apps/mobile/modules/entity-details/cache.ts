import type { QueryClient, QueryKey } from "@tanstack/react-query";

type EntityPatch = Record<string, unknown>;

export type EntityCacheSnapshot = {
  key: QueryKey;
  before: unknown;
  after: unknown;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const patchEntity = (
  value: unknown,
  entityId: string,
  patch: EntityPatch,
): unknown => {
  if (Array.isArray(value)) {
    const next = value.map((item) => patchEntity(item, entityId, patch));
    return next.some((item, index) => item !== value[index]) ? next : value;
  }
  if (!isRecord(value)) return value;

  let changed = value.id === entityId;
  let next: Record<string, unknown> = changed ? { ...value, ...patch } : value;
  for (const [key, child] of Object.entries(value)) {
    if (!Array.isArray(child) && !isRecord(child)) continue;
    const updated = patchEntity(child, entityId, patch);
    if (updated === child) continue;
    if (!changed) next = { ...value };
    next[key] = updated;
    changed = true;
  }
  return changed ? next : value;
};

export const optimisticallyPatchEntity = async (
  client: QueryClient,
  queryKeys: QueryKey[],
  entityId: string,
  patch: EntityPatch,
) => {
  await Promise.all(
    queryKeys.map((queryKey) => client.cancelQueries({ queryKey })),
  );
  const snapshots: EntityCacheSnapshot[] = [];
  for (const queryKey of queryKeys) {
    for (const [key, before] of client.getQueriesData({ queryKey })) {
      const after = patchEntity(before, entityId, patch);
      if (after === before) continue;
      client.setQueryData(key, after);
      snapshots.push({ key, before, after: client.getQueryData(key) });
    }
  }
  return snapshots;
};

export const restoreEntityCache = (
  client: QueryClient,
  snapshots: EntityCacheSnapshot[],
) => {
  for (const snapshot of snapshots) {
    // Preserve a newer optimistic or confirmed mutation on the same entity.
    if (client.getQueryData(snapshot.key) === snapshot.after)
      client.setQueryData(snapshot.key, snapshot.before);
  }
};
