import { ApiError, createApiClient, type RequestOptions } from "api-client";
import { getApiUrl } from "@/lib/api-url";
import { getRequestOptionsScope } from "./request-options-scope";

const apiURL = getApiUrl();
const MAYA_MAX_RETRY_AFTER_MS = 2_000;
const MAYA_RETRY_BACKOFF_LIMIT_MS = 1_000;

export type WorkspaceCtx = {
  session?: {
    token?: string;
  } | null;
  workspaceSlug: string;
  cookieHeader?: string;
};

const parseResponse = async <T>(response: Response) => {
  if (response.status === 204 || response.status === 205) {
    return { data: null } as T;
  }

  const text = await response.text();

  if (!text.trim()) {
    return { data: null } as T;
  }

  return JSON.parse(text) as T;
};

const createWorkspaceClient = (ctx: WorkspaceCtx) =>
  createApiClient(`${apiURL}/workspaces/${ctx.workspaceSlug}`);

const composeAbortSignals = (
  scopeSignal: AbortSignal,
  requestSignal?: AbortSignal | null,
) => {
  if (!requestSignal || requestSignal === scopeSignal) return scopeSignal;

  if (typeof AbortSignal.any === "function") {
    return AbortSignal.any([scopeSignal, requestSignal]);
  }

  const controller = new AbortController();
  const signals = [scopeSignal, requestSignal] as const;
  const abort = (signal: AbortSignal) => {
    for (const candidate of signals) {
      candidate.removeEventListener("abort", onAbort);
    }
    controller.abort(signal.reason);
  };
  const onAbort = (event: Event) => {
    abort(event.currentTarget as AbortSignal);
  };

  const abortedSignal = signals.find((signal) => signal.aborted);
  if (abortedSignal) {
    abort(abortedSignal);
  } else {
    for (const signal of signals) {
      signal.addEventListener("abort", onAbort, { once: true });
    }
  }

  return controller.signal;
};

const clampRetryDelay = (value: number | undefined, maximum: number) =>
  value === undefined ? maximum : Math.min(value, maximum);

const getMayaRetryOptions = (retry: RequestOptions["retry"]) => {
  if (retry === 0) return retry;

  const retryOptions = typeof retry === "number" ? { limit: retry } : retry;
  return {
    ...retryOptions,
    backoffLimit: clampRetryDelay(
      retryOptions?.backoffLimit,
      MAYA_RETRY_BACKOFF_LIMIT_MS,
    ),
    maxRetryAfter: clampRetryDelay(
      retryOptions?.maxRetryAfter,
      MAYA_MAX_RETRY_AFTER_MS,
    ),
  };
};

const getScopedRequestOptions = (options?: RequestOptions) => {
  const scope = getRequestOptionsScope();
  if (!scope) return options;

  return {
    ...options,
    retry: getMayaRetryOptions(options?.retry),
    signal: composeAbortSignals(scope.signal, options?.signal),
  } satisfies RequestOptions;
};

export { ApiError };

export const get = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const response = await createWorkspaceClient(ctx).get(
    url,
    getScopedRequestOptions(options),
  );
  return parseResponse<T>(response);
};

export const post = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const client = createWorkspaceClient(ctx);
  const requestOptions = getScopedRequestOptions(options);
  const response =
    json instanceof FormData
      ? await client.post(url, { body: json, ...requestOptions })
      : await client.post(url, { json, ...requestOptions });

  return parseResponse<U>(response);
};

export const put = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const client = createWorkspaceClient(ctx);
  const requestOptions = getScopedRequestOptions(options);
  const response =
    json instanceof FormData
      ? await client.put(url, { body: json, ...requestOptions })
      : await client.put(url, { json, ...requestOptions });

  return parseResponse<U>(response);
};

export const patch = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const client = createWorkspaceClient(ctx);
  const requestOptions = getScopedRequestOptions(options);
  const response =
    json instanceof FormData
      ? await client.patch(url, { body: json, ...requestOptions })
      : await client.patch(url, { json, ...requestOptions });

  return parseResponse<U>(response);
};

export const remove = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const response = await createWorkspaceClient(ctx).delete(
    url,
    getScopedRequestOptions(options),
  );
  return parseResponse<T>(response);
};
