/* eslint-disable no-await-in-loop -- Processing retries and bounded stream reads must execute sequentially. */
const MAX_ENTRIES = 64;
const MAX_BYTES = 16 * 1024 * 1024;
const MAX_THUMBNAIL_BYTES = 2 * 1024 * 1024;
export const THUMBNAIL_ACCESS_LEASE_MS = 30_000;

type CachedThumbnail = { blob: Blob; etag: string | null; checkedAt: number };
type PendingThumbnail = {
  controller: AbortController;
  consumers: Set<symbol>;
  promise: Promise<CachedThumbnail>;
};
const canceled = () =>
  new DOMException("Thumbnail request canceled", "AbortError");

const delay = (ms: number, signal: AbortSignal) =>
  new Promise<void>((resolve, reject) => {
    if (signal.aborted) {
      reject(canceled());
      return;
    }
    const cancel = () => {
      clearTimeout(timer);
      reject(canceled());
    };
    const timer = setTimeout(() => {
      signal.removeEventListener("abort", cancel);
      resolve();
    }, ms);
    signal.addEventListener("abort", cancel, { once: true });
  });

export const createThumbnailCache = () => {
  const entries = new Map<string, CachedThumbnail>();
  const pending = new Map<string, PendingThumbnail>();
  const listeners = new Set<() => void>();
  let totalBytes = 0;
  let revision = 0;
  let generation = 0;

  const remove = (url: string) => {
    const previous = entries.get(url);
    if (previous) totalBytes -= previous.blob.size;
    entries.delete(url);
  };
  const save = (url: string, entry: CachedThumbnail) => {
    remove(url);
    entries.set(url, entry);
    totalBytes += entry.blob.size;
    while (entries.size > MAX_ENTRIES || totalBytes > MAX_BYTES) {
      const oldest = entries.keys().next().value;
      if (!oldest) break;
      remove(oldest);
    }
  };
  const request = async (
    url: string,
    previous: CachedThumbnail | undefined,
    refresh: boolean,
    signal: AbortSignal,
  ) => {
    // Processing is asynchronous; retries are bounded and stop immediately when
    // virtualization removes the last consumer of this file.
    for (let attempt = 0; attempt < 12; attempt++) {
      const requestURL = refresh
        ? `${url}${url.includes("?") ? "&" : "?"}refresh=1`
        : url;
      const response = await fetch(requestURL, {
        cache: "no-store",
        credentials: "same-origin",
        signal,
        headers: previous?.etag
          ? { "If-None-Match": previous.etag }
          : undefined,
      });
      if (response.status === 202) {
        await response.body?.cancel();
        await delay(Math.min(2000 + attempt * 500, 5000), signal);
        continue;
      }
      if (response.status === 304 && previous)
        return { ...previous, checkedAt: Date.now() };
      if (
        !response.ok ||
        response.headers.get("content-type")?.split(";", 1)[0] !== "image/jpeg"
      ) {
        await response.body?.cancel();
        remove(url);
        throw new Error("Attachment preview unavailable");
      }
      const length = Number(response.headers.get("content-length"));
      if (length > MAX_THUMBNAIL_BYTES || !response.body) {
        await response.body?.cancel();
        throw new Error("Invalid thumbnail size");
      }
      const reader = response.body.getReader();
      const chunks: Uint8Array<ArrayBuffer>[] = [];
      let bytes = 0;
      try {
        for (;;) {
          const { value, done } = await reader.read();
          if (done) break;
          bytes += value.byteLength;
          if (bytes > MAX_THUMBNAIL_BYTES)
            throw new Error("Thumbnail size limit exceeded");
          chunks.push(new Uint8Array(value));
        }
      } catch (error) {
        await reader.cancel();
        throw error;
      } finally {
        reader.releaseLock();
      }
      if (!bytes) throw new Error("Empty thumbnail");
      return {
        blob: new Blob(chunks, { type: "image/jpeg" }),
        etag: response.headers.get("etag"),
        checkedAt: Date.now(),
      };
    }
    throw new Error("Thumbnail processing is taking longer than expected");
  };

  const load = (
    url: string,
    signal: AbortSignal,
    refresh = false,
    revalidate = false,
  ): Promise<CachedThumbnail> => {
    if (signal.aborted) return Promise.reject(canceled());
    const existing = entries.get(url);
    if (
      !refresh &&
      !revalidate &&
      existing &&
      Date.now() - existing.checkedAt < THUMBNAIL_ACCESS_LEASE_MS
    ) {
      save(url, existing);
      return Promise.resolve(existing);
    }
    let job = pending.get(url);
    if (!job) {
      const controller = new AbortController();
      const requestGeneration = generation;
      const promise = request(url, existing, refresh, controller.signal)
        .then((result) => {
          if (controller.signal.aborted || requestGeneration !== generation)
            throw canceled();
          save(url, result);
          return result;
        })
        .finally(() => {
          if (pending.get(url) === created) pending.delete(url);
        });
      const created: PendingThumbnail = {
        controller,
        consumers: new Set(),
        promise,
      };
      pending.set(url, created);
      job = created;
    }
    const active = job;
    const token = Symbol("thumbnail-consumer");
    active.consumers.add(token);
    return new Promise((resolve, reject) => {
      const release = () => {
        active.consumers.delete(token);
        signal.removeEventListener("abort", abort);
        if (!active.consumers.size && pending.get(url) === active) {
          pending.delete(url);
          active.controller.abort();
        }
      };
      const abort = () => {
        release();
        reject(canceled());
      };
      signal.addEventListener("abort", abort, { once: true });
      void active.promise.then(
        (result) => {
          release();
          if (!signal.aborted) resolve(result);
        },
        (error: unknown) => {
          release();
          reject(
            error instanceof Error
              ? error
              : new Error("Thumbnail request failed"),
          );
        },
      );
    });
  };
  const clear = () => {
    revision++;
    generation++;
    entries.clear();
    totalBytes = 0;
    for (const job of pending.values()) job.controller.abort();
    pending.clear();
    listeners.forEach((listener) => {
      listener();
    });
  };
  return {
    load,
    clear,
    refresh: async (url: string, signal: AbortSignal) => {
      const result = await load(url, signal, true);
      revision++;
      listeners.forEach((listener) => {
        listener();
      });
      return result.checkedAt;
    },
    revision: () => revision,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
  };
};

export type ThumbnailCache = ReturnType<typeof createThumbnailCache>;
