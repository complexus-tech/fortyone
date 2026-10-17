import { useSyncExternalStore } from "react";

const subscribeToHydration = () => () => undefined;
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;

// Persisted fields must not change the server-rendered first frame.
export const usePreviewCacheReady = () =>
  useSyncExternalStore(
    subscribeToHydration,
    getClientSnapshot,
    getServerSnapshot,
  );

/** Successful previews are small, optional browser cache entries, never full resources. */
export const PREVIEW_CACHE_TTL_MS = 24 * 60 * 60 * 1000;
const STORAGE_PREFIX = "fortyone:link-preview:v1:";
const MAX_ENTRIES = 100;
const MAX_ENTRY_BYTES = 16_000;

type PreviewCacheEntry<T> = { data: T; savedAt: number };

const parseEntry = (raw: string | null): PreviewCacheEntry<unknown> | null => {
  if (!raw || raw.length > MAX_ENTRY_BYTES) return null;
  const value: unknown = JSON.parse(raw);
  if (
    typeof value !== "object" ||
    value === null ||
    !("savedAt" in value) ||
    typeof value.savedAt !== "number" ||
    !Number.isFinite(value.savedAt) ||
    value.savedAt > Date.now() ||
    Date.now() - value.savedAt >= PREVIEW_CACHE_TTL_MS ||
    !("data" in value)
  )
    return null;
  return { data: value.data, savedAt: value.savedAt };
};

export const readPreviewCache = <T>(
  key: string,
  isValid: (value: unknown) => value is T,
): PreviewCacheEntry<T> | undefined => {
  try {
    const entry = parseEntry(localStorage.getItem(STORAGE_PREFIX + key));
    if (entry && isValid(entry.data))
      return { data: entry.data, savedAt: entry.savedAt };
    localStorage.removeItem(STORAGE_PREFIX + key);
  } catch {
    // Storage can be unavailable or malformed. The regular query remains usable.
  }
};

export const removePreviewCache = (key: string) => {
  try {
    localStorage.removeItem(STORAGE_PREFIX + key);
  } catch {
    // A disabled browser cache must not prevent refreshing the live preview.
  }
};

export const writePreviewCache = <T>(
  key: string,
  data: T,
  savedAt = Date.now(),
) => {
  try {
    const serialized = JSON.stringify({ data, savedAt });
    if (serialized.length > MAX_ENTRY_BYTES) return;
    const entries: { key: string; savedAt: number }[] = [];
    for (let index = 0; index < localStorage.length; index++) {
      const entryKey = localStorage.key(index);
      if (!entryKey?.startsWith(STORAGE_PREFIX)) continue;
      let entry: PreviewCacheEntry<unknown> | null = null;
      try {
        entry = parseEntry(localStorage.getItem(entryKey));
      } catch {
        // Malformed entries are evicted along with expired entries below.
      }
      entries.push({ key: entryKey, savedAt: entry?.savedAt ?? 0 });
    }
    const storageKey = STORAGE_PREFIX + key;
    entries.sort((first, second) => second.savedAt - first.savedAt);
    let retained = 0;
    for (const entry of entries) {
      if (entry.key === storageKey) continue;
      if (!entry.savedAt || retained >= MAX_ENTRIES - 1)
        localStorage.removeItem(entry.key);
      else retained++;
    }
    localStorage.setItem(storageKey, serialized);
  } catch {
    // Quota/private-mode failures fall back to the in-memory query cache.
  }
};
