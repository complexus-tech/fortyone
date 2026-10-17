import ky from "ky";
import type { LinkMetadata } from "./metadata";
import { normalizeMetadataUrl } from "./metadata";

export type { LinkMetadata } from "./metadata";

const MAX_CONCURRENT_METADATA_REQUESTS = 4;
const waitingRequests: (() => void)[] = [];
let activeRequests = 0;

const drainRequests = () => {
  while (
    activeRequests < MAX_CONCURRENT_METADATA_REQUESTS &&
    waitingRequests.length
  ) {
    waitingRequests.shift()?.();
  }
};

const abortReason = (signal: AbortSignal): Error => {
  const reason: unknown = signal.reason;
  if (reason instanceof Error) return reason;
  const error = new Error("Metadata request aborted");
  error.name = "AbortError";
  return error;
};

// Virtualized chat rows can mount together or disappear before a request starts.
// Bound outbound work while allowing each query to cancel independently.
const withMetadataSlot = <T>(
  request: () => Promise<T>,
  signal?: AbortSignal,
): Promise<T> =>
  new Promise((resolve, reject) => {
    const onAbort = () => {
      const index = waitingRequests.indexOf(start);
      if (index >= 0) waitingRequests.splice(index, 1);
      signal?.removeEventListener("abort", onAbort);
      if (signal) reject(abortReason(signal));
    };
    const start = () => {
      signal?.removeEventListener("abort", onAbort);
      if (signal?.aborted) {
        reject(abortReason(signal));
        return;
      }
      activeRequests++;
      void Promise.resolve()
        .then(request)
        .then(resolve, reject)
        .finally(() => {
          activeRequests--;
          drainRequests();
        });
    };
    if (signal?.aborted) {
      reject(abortReason(signal));
      return;
    }
    signal?.addEventListener("abort", onAbort, { once: true });
    waitingRequests.push(start);
    drainRequests();
  });

export const getLinkMetadata = async (
  url: string,
  signal?: AbortSignal,
  { refresh = false }: { refresh?: boolean } = {},
): Promise<LinkMetadata> => {
  try {
    const validUrl = normalizeMetadataUrl(url);
    if (!validUrl) return {};
    return await withMetadataSlot(
      () =>
        ky
          .get(
            `/api/metadata?url=${encodeURIComponent(validUrl)}${refresh ? "&refresh=1" : ""}`,
            {
              retry: 0,
              signal,
              timeout: 12_000,
            },
          )
          .json<LinkMetadata>(),
      signal,
    );
  } catch (error) {
    // Cancellation must not populate the cache with a failed preview. Ordinary
    // failures are optional enrichment: the original link remains usable.
    if (signal?.aborted) throw abortReason(signal);
    if (refresh) throw error;
    return {};
  }
};
