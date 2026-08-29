export interface RetryOptions {
  maxAttempts?: number;
  baseDelayMs?: number;
  maxDelayMs?: number;
}

interface NormalizedRetryOptions {
  maxAttempts: number;
  baseDelayMs: number;
  maxDelayMs: number;
}

const DEFAULT_RETRY_OPTIONS: NormalizedRetryOptions = {
  maxAttempts: 3,
  baseDelayMs: 250,
  maxDelayMs: 5_000,
};

export const createRetryingFetch = (
  fetchImplementation: typeof globalThis.fetch,
  options: RetryOptions = {},
): typeof globalThis.fetch => {
  const policy = normalizeRetryOptions(options);

  return async (input, init) => {
    const method = requestMethod(input, init);
    if (method !== "GET" && method !== "HEAD") {
      return fetchImplementation(input, init);
    }

    for (let attempt = 1; ; attempt += 1) {
      try {
        const response = await fetchImplementation(input, init);
        if (attempt >= policy.maxAttempts || !isRetryableStatus(response.status)) {
          return response;
        }
        const delay = responseDelay(
          response,
          policy,
          attempt,
          Date.now(),
          Math.random,
        );
        if (delay === undefined) {
          return response;
        }
        await response.body?.cancel().catch(() => undefined);
        await sleepWithSignal(delay, requestSignal(input, init));
      } catch (error) {
        if (
          attempt >= policy.maxAttempts ||
          requestSignal(input, init)?.aborted ||
          isAbortError(error)
        ) {
          throw error;
        }
        await sleepWithSignal(
          exponentialDelay(policy, attempt, Math.random),
          requestSignal(input, init),
        );
      }
    }
  };
};

const normalizeRetryOptions = (
  options: RetryOptions,
): NormalizedRetryOptions => {
  const normalized = { ...DEFAULT_RETRY_OPTIONS, ...options };
  if (
    !Number.isInteger(normalized.maxAttempts) ||
    normalized.maxAttempts < 1 ||
    normalized.maxAttempts > 10
  ) {
    throw new Error("retry maxAttempts must be an integer from 1 through 10");
  }
  if (
    !Number.isFinite(normalized.baseDelayMs) ||
    normalized.baseDelayMs < 1 ||
    !Number.isFinite(normalized.maxDelayMs) ||
    normalized.maxDelayMs < normalized.baseDelayMs ||
    normalized.maxDelayMs > 60_000
  ) {
    throw new Error(
      "retry delays must be positive, ordered, and no greater than 60 seconds",
    );
  }
  return normalized;
};

const requestMethod = (input: RequestInfo | URL, init?: RequestInit) =>
  (init?.method ?? (input instanceof Request ? input.method : "GET")).toUpperCase();

const requestSignal = (input: RequestInfo | URL, init?: RequestInit) =>
  init?.signal ?? (input instanceof Request ? input.signal : undefined);

const isRetryableStatus = (status: number) => status === 429 || status === 503;

const responseDelay = (
  response: Response,
  policy: NormalizedRetryOptions,
  attempt: number,
  now: number,
  random: () => number,
) => {
  const retryAfter = parseRetryAfter(response.headers.get("Retry-After"), now);
  if (retryAfter !== undefined) {
    return retryAfter <= policy.maxDelayMs ? retryAfter : undefined;
  }
  return exponentialDelay(policy, attempt, random);
};

const parseRetryAfter = (value: string | null, now: number) => {
  if (value === null) {
    return undefined;
  }
  const trimmed = value.trim();
  if (/^\d+$/u.test(trimmed)) {
    return Number(trimmed) * 1_000;
  }
  const parsed = Date.parse(trimmed);
  return Number.isNaN(parsed) ? undefined : Math.max(0, parsed - now);
};

const exponentialDelay = (
  policy: NormalizedRetryOptions,
  attempt: number,
  random: () => number,
) => {
  const ceiling = Math.min(
    policy.maxDelayMs,
    policy.baseDelayMs * 2 ** (attempt - 1),
  );
  return Math.floor(Math.max(0, Math.min(1, random())) * ceiling);
};

const isAbortError = (error: unknown) =>
  error instanceof DOMException && error.name === "AbortError";

const sleepWithSignal = (milliseconds: number, signal?: AbortSignal | null) =>
  new Promise<void>((resolve, reject) => {
    if (signal?.aborted) {
      reject(signal.reason);
      return;
    }
    const onAbort = () => {
      clearTimeout(timer);
      reject(signal?.reason);
    };
    const timer = setTimeout(() => {
      signal?.removeEventListener("abort", onAbort);
      resolve();
    }, milliseconds);
    signal?.addEventListener("abort", onAbort, { once: true });
  });
