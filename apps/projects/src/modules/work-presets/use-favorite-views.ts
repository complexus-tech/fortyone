"use client";

import { useMemo } from "react";
import { useFavorites } from "@/shared/favorites";
import type { FavoriteViewRef } from "./favorites-store";

export const useFavoriteViews = () => {
  const store = useFavorites();
  const favorites = useMemo(
    () =>
      store.favorites.flatMap((item) =>
        item.kind === "view" ? [{ id: item.id, teamId: item.teamId }] : [],
      ),
    [store.favorites],
  );
  return {
    favorites,
    enabled: store.enabled,
    isFavorite: (view: FavoriteViewRef) =>
      store.isFavorite({ kind: "view", ...view }),
    toggleFavorite: (view: FavoriteViewRef) => {
      store.toggleFavorite({ kind: "view", ...view });
    },
    removeFavorite: (view: FavoriteViewRef) => {
      store.removeFavorite({ kind: "view", ...view });
    },
  };
};
