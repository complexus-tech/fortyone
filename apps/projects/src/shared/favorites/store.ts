import { z } from "zod";
import type { FavoriteRef, FavoriteScope } from "./types";
import { favoriteIdentity, favoriteRefSchema } from "./types";

export type FavoriteSnapshot = {
  favorites: readonly FavoriteRef[];
  error: string | null;
};
const EMPTY: FavoriteSnapshot = { favorites: [], error: null };
const UNAVAILABLE: FavoriteSnapshot = {
  favorites: [],
  error: "Favorites could not be loaded in this browser.",
};
const INVALID: FavoriteSnapshot = {
  favorites: [],
  error: "Saved favorites could not be read. Add a favorite to start again.",
};
const STORE_EVENT = "fortyone:favorites";
const LEGACY_EVENT = "fortyone:view-favorites";
const LEGACY_PREFIX = "work-presets:favorites:v1:";
const schema = z.object({
  version: z.literal(2),
  favorites: z.array(favoriteRefSchema),
});
const legacySchema = z.array(z.object({ id: z.uuid(), teamId: z.uuid() }));
const snapshots = new Map<
  string,
  { raw: string | null; legacy: boolean; value: FavoriteSnapshot }
>();
const subscriptions = new Map<
  string,
  { scope: FavoriteScope; listeners: Set<() => void> }
>();

export const favoritesKey = ({ workspaceSlug, userId }: FavoriteScope) =>
  `fortyone:favorites:v2:${encodeURIComponent(workspaceSlug)}:${encodeURIComponent(userId)}`;
export const legacyViewFavoritesKey = ({
  workspaceSlug,
  userId,
}: FavoriteScope) =>
  `${LEGACY_PREFIX}${encodeURIComponent(workspaceSlug)}:${encodeURIComponent(userId)}`;

export const scopeFromLegacyKey = (key: string): FavoriteScope | null => {
  if (!key.startsWith(LEGACY_PREFIX)) return null;
  const parts = key.slice(LEGACY_PREFIX.length).split(":");
  if (parts.length !== 2 || !parts.every(Boolean)) return null;
  try {
    return {
      workspaceSlug: decodeURIComponent(parts[0]),
      userId: decodeURIComponent(parts[1]),
    };
  } catch {
    return null;
  }
};

const uniqueFavorites = (items: FavoriteRef[]) => {
  const seen = new Set<string>();
  return items.filter((item) => {
    const identity = favoriteIdentity(item);
    if (seen.has(identity)) return false;
    seen.add(identity);
    return true;
  });
};

export const getServerFavoriteSnapshot = () => EMPTY;

/** Legacy preferences are projected on read and migrated on the next successful write. */
export const readFavoriteSnapshot = (
  scope: FavoriteScope | null,
): FavoriteSnapshot => {
  if (!scope || typeof window === "undefined") return EMPTY;
  const key = favoritesKey(scope);
  let raw: string | null;
  let legacy = false;
  try {
    raw = window.localStorage.getItem(key);
    if (raw === null) {
      raw = window.localStorage.getItem(legacyViewFavoritesKey(scope));
      legacy = true;
    }
  } catch {
    return UNAVAILABLE;
  }
  const previous = snapshots.get(key);
  if (previous?.raw === raw && previous.legacy === legacy)
    return previous.value;
  let value = EMPTY;
  if (raw) {
    try {
      const parsed: unknown = JSON.parse(raw);
      const favorites: FavoriteRef[] = legacy
        ? legacySchema.parse(parsed).map((item) => ({ kind: "view", ...item }))
        : schema.parse(parsed).favorites;
      value = { favorites: uniqueFavorites(favorites), error: null };
    } catch {
      value = INVALID;
    }
  }
  snapshots.set(key, { raw, legacy, value });
  return value;
};

const notify = (key: string | null) => {
  for (const [storeKey, subscription] of subscriptions) {
    if (
      key === null ||
      key === storeKey ||
      key === legacyViewFavoritesKey(subscription.scope)
    )
      subscription.listeners.forEach((listener) => {
        listener();
      });
  }
};
const onStorage = (event: StorageEvent) => {
  notify(event.key);
};
const onChange = (event: Event) => {
  notify((event as CustomEvent<string>).detail);
};

/** All consumers share browser listeners, including legacy view adapters. */
export const subscribeFavorites = (
  scope: FavoriteScope | null,
  listener: () => void,
) => {
  if (!scope || typeof window === "undefined") return () => {};
  if (!subscriptions.size) {
    window.addEventListener("storage", onStorage);
    window.addEventListener(STORE_EVENT, onChange);
    window.addEventListener(LEGACY_EVENT, onChange);
  }
  const key = favoritesKey(scope);
  let subscription = subscriptions.get(key);
  if (!subscription) {
    subscription = { scope, listeners: new Set() };
    subscriptions.set(key, subscription);
  }
  subscription.listeners.add(listener);
  return () => {
    subscription.listeners.delete(listener);
    if (!subscription.listeners.size) subscriptions.delete(key);
    if (!subscriptions.size) {
      window.removeEventListener("storage", onStorage);
      window.removeEventListener(STORE_EVENT, onChange);
      window.removeEventListener(LEGACY_EVENT, onChange);
    }
  };
};

export const updateFavorites = (
  scope: FavoriteScope,
  update: (current: readonly FavoriteRef[]) => FavoriteRef[],
) => {
  const current = readFavoriteSnapshot(scope);
  if (current === UNAVAILABLE) throw new Error(current.error!);
  const favorites = uniqueFavorites(
    z.array(favoriteRefSchema).parse(update(current.favorites)),
  );
  window.localStorage.setItem(
    favoritesKey(scope),
    JSON.stringify({ version: 2, favorites }),
  );
  // Retain v1 as recovery data; v2 takes precedence, including an explicitly empty list.
  window.dispatchEvent(
    new CustomEvent(STORE_EVENT, { detail: favoritesKey(scope) }),
  );
};
