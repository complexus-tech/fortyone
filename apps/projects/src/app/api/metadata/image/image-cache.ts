/* eslint-disable no-await-in-loop -- Poll a lease owner sequentially without overlapping object-store requests. */
import { randomUUID } from "node:crypto";
import type { SharedPreviewCache } from "../shared-cache";
import { previewCacheKey, sharedPreviewCache } from "../shared-cache";
import { MetadataCapacityError } from "../metadata-cache";
import { fetchPublicImage } from "../safe-fetch";
import { parseMetadataUrl } from "../target-policy";
import type { PreviewImage, PreviewImageStore } from "./object-store";
import { previewImageStore } from "./object-store";

const IMAGE_TTL_MS = 24 * 60 * 60 * 1000;
const MAX_LOCAL_BYTES = 32 * 1024 * 1024;
const MAX_IN_FLIGHT = 8;

export const createImageCache = ({
  load = (url, signal) => fetchPublicImage(url, { signal }),
  storage,
  shared,
  now = Date.now,
}: {
  load?: (url: URL, signal: AbortSignal) => Promise<PreviewImage>;
  storage?: PreviewImageStore;
  shared?: SharedPreviewCache;
  now?: () => number;
} = {}) => {
  const entries = new Map<string, { image: PreviewImage; expiresAt: number }>();
  const pending = new Map<string, Promise<PreviewImage>>();
  const failures = new Map<string, number>();
  let totalBytes = 0;
  const remember = (key: string, image: PreviewImage) => {
    totalBytes -= entries.get(key)?.image.body.byteLength ?? 0;
    entries.delete(key);
    entries.set(key, { image, expiresAt: now() + IMAGE_TTL_MS });
    totalBytes += image.body.byteLength;
    while (totalBytes > MAX_LOCAL_BYTES || entries.size > 256) {
      const oldest = entries.keys().next().value;
      if (!oldest) break;
      totalBytes -= entries.get(oldest)!.image.body.byteLength;
      entries.delete(oldest);
    }
  };
  const get = async (input: string, version: string, signal?: AbortSignal) => {
    signal?.throwIfAborted();
    const url = parseMetadataUrl(input);
    url.hash = "";
    if (url.href.length > 8192) throw new Error("Image URL is too long.");
    const key = previewCacheKey(`${url.href}\n${version}`);
    const cached = entries.get(key);
    if (cached && cached.expiresAt > now()) {
      entries.delete(key);
      entries.set(key, cached);
      return cached.image;
    }
    if ((failures.get(key) ?? 0) > now())
      throw new Error("Preview image temporarily unavailable.");
    let request = pending.get(key);
    if (!request) {
      if (pending.size >= MAX_IN_FLIGHT) throw new MetadataCapacityError();
      request = (async () => {
        const token = randomUUID();
        let ownsLock = false;
        try {
          const stored = await storage?.read(key);
          if (stored) {
            remember(key, stored);
            return stored;
          }
          const sharedStorageAvailable =
            storage && (!storage.available || storage.available());
          const claim = sharedStorageAvailable
            ? await shared?.claim(`image-lock:${key}`, token, 15_000)
            : null;
          ownsLock = claim === true;
          if (claim === false && storage) {
            for (let attempt = 0; attempt < 10; attempt++) {
              if (storage.available && !storage.available()) break;
              await new Promise((resolve) => {
                setTimeout(resolve, 750);
              });
              const ready = await storage.read(key);
              if (ready) {
                remember(key, ready);
                return ready;
              }
            }
            // A failed writer or unavailable object store must not leave the
            // image broken. Per-process admission still bounds the fallback.
          }
          // Work is shared by several img elements. It has its own bounded
          // deadline so one scrolling-away client cannot cancel their peers.
          const image = await load(url, AbortSignal.timeout(8_000));
          remember(key, image);
          await storage?.write(key, image);
          return image;
        } catch (error) {
          if (!(error instanceof MetadataCapacityError)) {
            failures.delete(key);
            failures.set(key, now() + 30_000);
            while (failures.size > 256)
              failures.delete(failures.keys().next().value!);
          }
          throw error;
        } finally {
          pending.delete(key);
          if (ownsLock) await shared?.release(`image-lock:${key}`, token);
        }
      })();
      pending.set(key, request);
    }
    // Abort only this subscriber; coalesced upstream work remains bounded.
    if (!signal) return request;
    return new Promise<PreviewImage>((resolve, reject) => {
      const onAbort = () => {
        reject(
          signal.reason instanceof Error
            ? signal.reason
            : new DOMException("Aborted", "AbortError"),
        );
      };
      signal.addEventListener("abort", onAbort, { once: true });
      request.then(resolve, reject).finally(() => {
        signal.removeEventListener("abort", onAbort);
      });
      if (signal.aborted) onAbort();
    });
  };
  return {
    get,
    clear: () => {
      entries.clear();
      failures.clear();
      totalBytes = 0;
    },
  };
};

export const metadataImageCache = createImageCache({
  storage: previewImageStore,
  shared: sharedPreviewCache,
});
