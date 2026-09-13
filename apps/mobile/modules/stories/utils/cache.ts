import type { QueryClient, QueryKey } from "@tanstack/react-query";
import type { DetailedStory } from "../types";

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const mapChanged = (values: unknown[], update: (value: unknown) => unknown) => {
  const next = values.map(update);
  return next.some((value, index) => value !== values[index]) ? next : values;
};

/** Preserve cache envelopes and identity for entries that were not changed. */
export const updateStoryCache = (
  data: unknown,
  storyId: string,
  patch: Partial<DetailedStory>,
): unknown => {
  const update = (value: unknown): unknown => {
    if (Array.isArray(value)) return mapChanged(value, update);
    if (!isRecord(value)) return value;
    if (typeof value.id === "string" && typeof value.title === "string") {
      const subStories = Array.isArray(value.subStories)
        ? mapChanged(value.subStories, update)
        : value.subStories;
      if (value.id === storyId) {
        return {
          ...value,
          ...(subStories !== value.subStories ? { subStories } : {}),
          ...patch,
        };
      }
      return subStories !== value.subStories ? { ...value, subStories } : value;
    }
    for (const field of ["groups", "pages", "stories"] as const) {
      const values = value[field];
      if (!Array.isArray(values)) continue;
      const next = mapChanged(values, update);
      return next === values ? value : { ...value, [field]: next };
    }
    return value;
  };
  return update(data);
};

export type StoryCacheSnapshot = {
  key: QueryKey;
  before: unknown;
  after: unknown;
};

export const optimisticallyUpdateStory = async (
  client: QueryClient,
  queryKey: QueryKey,
  storyId: string,
  patch: Partial<DetailedStory>,
) => optimisticallyUpdateStories(client, queryKey, [{ storyId, patch }]);

export const optimisticallyUpdateStories = async (
  client: QueryClient,
  queryKey: QueryKey,
  updates: { storyId: string; patch: Partial<DetailedStory> }[],
) => {
  await client.cancelQueries({ queryKey });
  const snapshots: StoryCacheSnapshot[] = [];
  for (const [key, before] of client.getQueriesData({ queryKey })) {
    const after = updates.reduce(
      (data, update) => updateStoryCache(data, update.storyId, update.patch),
      before,
    );
    if (after === before) continue;
    client.setQueryData(key, after);
    snapshots.push({ key, before, after: client.getQueryData(key) });
  }
  return snapshots;
};

export const restoreStoryCache = (
  client: QueryClient,
  snapshots: StoryCacheSnapshot[],
) => {
  for (const snapshot of snapshots) {
    if (client.getQueryData(snapshot.key) === snapshot.after) {
      client.setQueryData(snapshot.key, snapshot.before);
    }
  }
};
