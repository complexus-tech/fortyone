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

export type ResponseDecoder<T> = (value: unknown) => T;

export class ApiContractError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiContractError";
    this.status = status;
  }
}

const decodeResponseValue = <T>(
  value: unknown,
  status: number,
  decode?: ResponseDecoder<T>,
) => {
  if (!decode) return value as T;

  try {
    return decode(value);
  } catch (cause) {
    if (cause instanceof ApiContractError) throw cause;

    // Runtime decoders may include response values in their errors. Do not
    // retain that cause at this transport boundary where monitoring captures it.
    throw new ApiContractError(
      `API response did not match its runtime contract (${status})`,
      status,
    );
  }
};

const parseResponse = async <T>(
  response: Response,
  decode?: ResponseDecoder<T>,
) => {
  if (response.status === 204 || response.status === 205) {
    return decodeResponseValue({ data: null }, response.status, decode);
  }

  const text = await response.text();

  if (!text.trim()) {
    throw new ApiContractError(
      `API returned an empty ${response.status} response`,
      response.status,
    );
  }

  let value: unknown;
  try {
    value = JSON.parse(text) as unknown;
  } catch {
    throw new ApiContractError(
      `API returned invalid JSON with status ${response.status}`,
      response.status,
    );
  }

  return decodeResponseValue(value, response.status, decode);
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
  decode?: ResponseDecoder<T>,
) => {
  const response = await createWorkspaceClient(ctx).get(
    url,
    getScopedRequestOptions(options),
  );
  return parseResponse(response, decode);
};

export const post = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
  decode?: ResponseDecoder<U>,
) => {
  const client = createWorkspaceClient(ctx);
  const requestOptions = getScopedRequestOptions(options);
  const response =
    json instanceof FormData
      ? await client.post(url, { body: json, ...requestOptions })
      : await client.post(url, { json, ...requestOptions });

  return parseResponse(response, decode);
};

export const put = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
  decode?: ResponseDecoder<U>,
) => {
  const client = createWorkspaceClient(ctx);
  const requestOptions = getScopedRequestOptions(options);
  const response =
    json instanceof FormData
      ? await client.put(url, { body: json, ...requestOptions })
      : await client.put(url, { json, ...requestOptions });

  return parseResponse(response, decode);
};

export const patch = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
  decode?: ResponseDecoder<U>,
) => {
  const client = createWorkspaceClient(ctx);
  const requestOptions = getScopedRequestOptions(options);
  const response =
    json instanceof FormData
      ? await client.patch(url, { body: json, ...requestOptions })
      : await client.patch(url, { json, ...requestOptions });

  return parseResponse(response, decode);
};

export const remove = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
  decode?: ResponseDecoder<T>,
) => {
  const response = await createWorkspaceClient(ctx).delete(
    url,
    getScopedRequestOptions(options),
  );
  return parseResponse(response, decode);
};
