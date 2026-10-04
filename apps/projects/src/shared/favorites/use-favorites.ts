"use client";

import { useCallback, useSyncExternalStore } from "react";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { FavoriteRef } from "./types";
import { favoriteIdentity, sameFavorite } from "./types";
import {
  getServerFavoriteSnapshot,
  readFavoriteSnapshot,
  subscribeFavorites,
  updateFavorites,
} from "./store";

export const useFavorites = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const userId = session?.user.id;
  const scope = workspaceSlug && userId ? { workspaceSlug, userId } : null;
  const subscribe = useCallback(
    (listener: () => void) =>
      subscribeFavorites(
        workspaceSlug && userId ? { workspaceSlug, userId } : null,
        listener,
      ),
    [workspaceSlug, userId],
  );
  const getSnapshot = useCallback(
    () =>
      readFavoriteSnapshot(
        workspaceSlug && userId ? { workspaceSlug, userId } : null,
      ),
    [workspaceSlug, userId],
  );
  const snapshot = useSyncExternalStore(
    subscribe,
    getSnapshot,
    getServerFavoriteSnapshot,
  );
  const change = (
    update: (current: readonly FavoriteRef[]) => FavoriteRef[],
  ) => {
    if (!scope) return false;
    try {
      updateFavorites(scope, update);
      return true;
    } catch {
      toast.error("Favorites could not be saved in this browser.");
      return false;
    }
  };
  return {
    favorites: snapshot.favorites,
    error: snapshot.error,
    enabled: Boolean(scope),
    isFavorite: (item: FavoriteRef) =>
      snapshot.favorites.some((favorite) => sameFavorite(favorite, item)),
    toggleFavorite: (item: FavoriteRef) =>
      change((current) =>
        current.some((favorite) => sameFavorite(favorite, item))
          ? current.filter((favorite) => !sameFavorite(favorite, item))
          : [...current, item],
      ),
    removeFavorite: (item: FavoriteRef) =>
      change((current) =>
        current.filter((favorite) => !sameFavorite(favorite, item)),
      ),
    removeFavorites: (items: readonly FavoriteRef[]) => {
      const removing = new Set(items.map(favoriteIdentity));
      return change((current) =>
        current.filter((favorite) => !removing.has(favoriteIdentity(favorite))),
      );
    },
  };
};
