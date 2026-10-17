/* eslint-disable no-await-in-loop -- Wait for a distributed lease owner with bounded sequential polling. */
import { randomUUID } from "node:crypto";
import type { LinkMetadata } from "@/lib/queries/links/metadata";
import {
  hasLinkMetadata,
  METADATA_FAILURE_TTL_MS,
  METADATA_SUCCESS_TTL_MS,
} from "@/lib/queries/links/metadata";
import type { SharedPreviewCache } from "./shared-cache";
import { previewCacheKey } from "./shared-cache";
import { MetadataFetchError, parseMetadataUrl } from "./target-policy";

const RETENTION_MS = 24 * 60 * 60 * 1000;
const LOCK_TTL_MS = 12_000;
type CacheEntry = {
  freshUntil: number;
  expiresAt: number;
  metadata?: LinkMetadata;
  error?: { code: MetadataFetchError["code"]; message: string };
};
type PendingMetadata = {
  controller: AbortController;
  promise: Promise<LinkMetadata>;
  subscribers: number;
  settled: boolean;
};
type MetadataLoader = (url: URL, signal: AbortSignal) => Promise<LinkMetadata>;

export class MetadataCapacityError extends Error {
  constructor() {
    super("Metadata service is busy. Please try again shortly.");
    this.name = "MetadataCapacityError";
  }
}
const abortError = (signal: AbortSignal) =>
  signal.reason instanceof Error
    ? signal.reason
    : new DOMException("Metadata request aborted", "AbortError");

const readEntry = (value: string | null): CacheEntry | undefined => {
  if (!value || value.length > 32_768) return undefined;
  try {
    const entry: CacheEntry = JSON.parse(value);
    if (
      Number.isFinite(entry.freshUntil) &&
      Number.isFinite(entry.expiresAt) &&
      (entry.metadata || entry.error)
    )
      return entry;
  } catch {
    // An old schema or corrupt cache record is a miss, never authoritative data.
  }
  return undefined;
};

const wait = (milliseconds: number, signal: AbortSignal) =>
  new Promise<void>((resolve, reject) => {
    const onAbort = () => {
      clearTimeout(timer);
      reject(abortError(signal));
    };
    const timer = setTimeout(() => {
      signal.removeEventListener("abort", onAbort);
      resolve();
    }, milliseconds);
    signal.addEventListener("abort", onAbort, { once: true });
    if (signal.aborted) onAbort();
  });

// Stores public-page metadata only. Every route still authenticates before
// looking up the cache; every miss passes through the pinned-DNS fetcher.
export const createMetadataCache = (
  load: MetadataLoader,
  {
    maxEntries = 1000,
    maxInFlight = 32,
    now = Date.now,
    shared,
    schedule = (work) => {
      void work();
    },
  }: {
    maxEntries?: number;
    maxInFlight?: number;
    now?: () => number;
    shared?: SharedPreviewCache;
    schedule?: (work: () => Promise<void>) => void;
  } = {},
) => {
  const entries = new Map<string, CacheEntry>();
  const pending = new Map<string, PendingMetadata>();
  const remember = (key: string, entry: CacheEntry) => {
    entries.delete(key);
    entries.set(key, entry);
    while (entries.size > maxEntries) {
      const oldest = entries.keys().next().value;
      if (oldest === undefined) break;
      entries.delete(oldest);
    }
  };
  const publish = async (key: string, entry: CacheEntry) => {
    remember(key, entry);
    await shared?.write(
      `metadata:${key}`,
      JSON.stringify(entry),
      Math.max(1, entry.expiresAt - now()),
    );
  };
  const subscribe = (
    key: string,
    flight: PendingMetadata,
    signal?: AbortSignal,
  ) =>
    new Promise<LinkMetadata>((resolve, reject) => {
      let released = false;
      flight.subscribers++;
      const release = () => {
        if (released) return;
        released = true;
        signal?.removeEventListener("abort", onAbort);
        flight.subscribers--;
        if (!flight.subscribers && !flight.settled) {
          if (pending.get(key) === flight) pending.delete(key);
          flight.controller.abort();
        }
      };
      const onAbort = () => {
        if (signal) reject(abortError(signal));
        release();
      };
      if (signal?.aborted) {
        onAbort();
        return;
      }
      signal?.addEventListener("abort", onAbort, { once: true });
      flight.promise.then(
        (metadata) => {
          release();
          resolve(metadata);
        },
        (error: unknown) => {
          release();
          reject(
            error instanceof Error
              ? error
              : new Error("Metadata fetch failed.", { cause: error }),
          );
        },
      );
    });

  const refresh = (
    url: URL,
    key: string,
    previous?: CacheEntry,
    signal?: AbortSignal,
  ) => {
    const existing = pending.get(key);
    if (existing) return subscribe(key, existing, signal);
    if (pending.size >= maxInFlight)
      return Promise.reject(new MetadataCapacityError());
    const controller = new AbortController();
    const flight: PendingMetadata = {
      controller,
      promise: Promise.resolve({}),
      subscribers: 0,
      settled: false,
    };
    const token = randomUUID();
    let ownsLock = false;
    flight.promise = Promise.resolve()
      .then(async () => {
        if (controller.signal.aborted) throw abortError(controller.signal);
        const claim = shared
          ? await shared.claim(`metadata-lock:${key}`, token, LOCK_TTL_MS)
          : null;
        ownsLock = claim === true;
        if (claim === false) {
          // Another replica is already refreshing. Wait for its new revision rather
          // than launching the same upstream request in every process.
          for (let attempt = 0; attempt < 45; attempt++) {
            await wait(200, controller.signal);
            const entry = readEntry(await shared!.read(`metadata:${key}`));
            if (
              entry &&
              entry.freshUntil > now() &&
              (entry.metadata?.version !== previous?.metadata?.version ||
                entry.freshUntil !== previous?.freshUntil)
            ) {
              remember(key, entry);
              if (entry.error)
                throw new MetadataFetchError(
                  entry.error.code,
                  entry.error.message,
                );
              return entry.metadata ?? {};
            }
          }
          throw new MetadataCapacityError();
        }
        const metadata = await load(url, controller.signal);
        // The signal can change while awaiting the upstream request.
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- TypeScript retains an earlier narrowing across await.
        if (controller.signal.aborted) throw abortError(controller.signal);
        const successful = hasLinkMetadata(metadata);
        const entry: CacheEntry = {
          freshUntil:
            now() +
            (successful ? METADATA_SUCCESS_TTL_MS : METADATA_FAILURE_TTL_MS),
          expiresAt:
            now() + (successful ? RETENTION_MS : METADATA_FAILURE_TTL_MS),
          metadata,
        };
        await publish(key, entry);
        return metadata;
      })
      .catch(async (error: unknown) => {
        if (controller.signal.aborted) throw abortError(controller.signal);
        if (error instanceof MetadataCapacityError) throw error;
        const failure =
          error instanceof MetadataFetchError
            ? error
            : new MetadataFetchError(
                "upstream-error",
                "Metadata fetch failed.",
              );
        // Keep last-good content when a site is temporarily down. A manual
        // refresh still reports the error; ordinary readers receive stale data.
        if (
          previous?.metadata &&
          hasLinkMetadata(previous.metadata) &&
          previous.expiresAt > now()
        ) {
          await publish(key, {
            ...previous,
            freshUntil: now() + METADATA_FAILURE_TTL_MS,
          });
        } else {
          await publish(key, {
            freshUntil: now() + METADATA_FAILURE_TTL_MS,
            expiresAt: now() + METADATA_FAILURE_TTL_MS,
            error: { code: failure.code, message: failure.message },
          });
        }
        throw failure;
      })
      .finally(async () => {
        flight.settled = true;
        if (pending.get(key) === flight) pending.delete(key);
        if (ownsLock) await shared?.release(`metadata-lock:${key}`, token);
      });
    pending.set(key, flight);
    return subscribe(key, flight, signal);
  };

  const get = async (
    input: string | URL,
    signal?: AbortSignal,
    { force = false }: { force?: boolean } = {},
  ) => {
    if (signal?.aborted) throw abortError(signal);
    const url = parseMetadataUrl(input);
    url.hash = "";
    if (url.href.length > 8192)
      throw new MetadataFetchError("invalid-url", "Metadata URL is too long.");
    const key = previewCacheKey(url.href);
    // A shared read observes refreshes from other replicas. Browser query caching
    // prevents this round trip during repeated virtualized row mounts.
    const sharedEntry = shared
      ? readEntry(await shared.read(`metadata:${key}`))
      : undefined;
    const cached = sharedEntry ?? entries.get(key);
    if (cached && cached.expiresAt > now()) {
      remember(key, cached);
      if (!force && cached.freshUntil > now()) {
        if (cached.error)
          throw new MetadataFetchError(cached.error.code, cached.error.message);
        return cached.metadata ?? {};
      }
      if (!force && cached.metadata && hasLinkMetadata(cached.metadata)) {
        schedule(async () => {
          await refresh(url, key, cached).catch(() => undefined);
        });
        return { ...cached.metadata, stale: true };
      }
    }
    return refresh(url, key, cached, signal);
  };
  return {
    get,
    clear: () => {
      entries.clear();
      for (const flight of pending.values()) flight.controller.abort();
      pending.clear();
    },
  };
};
