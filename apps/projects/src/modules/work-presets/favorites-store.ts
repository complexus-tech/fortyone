import type { FavoriteRef } from "@/shared/favorites/types";
import {
  getServerFavoriteSnapshot,
  legacyViewFavoritesKey,
  readFavoriteSnapshot,
  scopeFromLegacyKey,
  subscribeFavorites,
  updateFavorites,
} from "@/shared/favorites/store";

export type FavoriteViewRef = { id: string; teamId: string };
const projections = new WeakMap<
  readonly FavoriteRef[],
  readonly FavoriteViewRef[]
>();
const project = (items: readonly FavoriteRef[]): readonly FavoriteViewRef[] => {
  const cached = projections.get(items);
  if (cached) return cached;
  const views = items.flatMap((item) =>
    item.kind === "view" ? [{ id: item.id, teamId: item.teamId }] : [],
  );
  projections.set(items, views);
  return views;
};

/** Retain the legacy key API while reads and writes use the generalized store. */
export const viewFavoritesKey = (workspace: string, userId: string) =>
  legacyViewFavoritesKey({ workspaceSlug: workspace, userId });
export const getServerFavorites = () =>
  project(getServerFavoriteSnapshot().favorites);
export const readViewFavorites = (key: string | null) =>
  project(readFavoriteSnapshot(key ? scopeFromLegacyKey(key) : null).favorites);
export const subscribeViewFavorites = (
  key: string | null,
  listener: () => void,
) => subscribeFavorites(key ? scopeFromLegacyKey(key) : null, listener);
export const updateViewFavorites = (
  key: string,
  update: (current: readonly FavoriteViewRef[]) => FavoriteViewRef[],
) => {
  const scope = scopeFromLegacyKey(key);
  if (!scope) throw new Error("Invalid favorites scope");
  updateFavorites(scope, (current) => [
    ...current.filter((item) => item.kind !== "view"),
    ...update(project(current)).map((item) => ({
      kind: "view" as const,
      ...item,
    })),
  ]);
};
