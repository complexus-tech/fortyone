import { ApiError, createApiClient, type RequestOptions } from "api-client";
import { getApiUrl } from "@/lib/api-url";

const apiURL = getApiUrl();

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

export { ApiError };

export const get = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const response = await createWorkspaceClient(ctx).get(url, options);
  return parseResponse<T>(response);
};

export const post = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const client = createWorkspaceClient(ctx);
  const response =
    json instanceof FormData
      ? await client.post(url, { body: json, ...options })
      : await client.post(url, { json, ...options });

  return parseResponse<U>(response);
};

export const put = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const client = createWorkspaceClient(ctx);
  const response =
    json instanceof FormData
      ? await client.put(url, { body: json, ...options })
      : await client.put(url, { json, ...options });

  return parseResponse<U>(response);
};

export const patch = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const client = createWorkspaceClient(ctx);
  const response =
    json instanceof FormData
      ? await client.patch(url, { body: json, ...options })
      : await client.patch(url, { json, ...options });

  return parseResponse<U>(response);
};

export const remove = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: RequestOptions,
) => {
  const response = await createWorkspaceClient(ctx).delete(url, options);
  return parseResponse<T>(response);
};
