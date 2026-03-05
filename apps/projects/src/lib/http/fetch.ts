import type { Options } from "ky";
import ky from "ky";
import type { Session } from "next-auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { ApiError } from "./error";

const apiURL = getApiUrl();

export type WorkspaceCtx = {
  session: Session;
  workspaceSlug: string;
  cookieHeader?: string;
};

const createClient = (ctx: WorkspaceCtx) => {
  const prefixUrl = `${apiURL}/workspaces/${ctx.workspaceSlug}/`;

  const headers = buildAuthHeaders({
    token: ctx.session?.token,
    cookieHeader: ctx.cookieHeader,
  });

  const client = ky.create({
    prefixUrl,
    credentials: "include",
    headers,
    hooks: {
      beforeError: [
        async (error) => {
          const { response } = error;
          if (response.body) {
            const data = await response.json();
            throw new ApiError(error.message, response.status, data);
          }
          throw error;
        },
      ],
    },
  });

  return client;
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

export const get = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);
  const response = await client.get(url, options);
  return parseResponse<T>(response);
};

export const post = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);

  if (json instanceof FormData) {
    const response = await client.post(url, { body: json, ...options });
    return parseResponse<U>(response);
  }

  const response = await client.post(url, { json, ...options });
  return parseResponse<U>(response);
};

export const put = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);

  if (json instanceof FormData) {
    const response = await client.put(url, { body: json, ...options });
    return parseResponse<U>(response);
  }

  const response = await client.put(url, { json, ...options });
  return parseResponse<U>(response);
};

export const patch = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);

  if (json instanceof FormData) {
    const response = await client.patch(url, { body: json, ...options });
    return parseResponse<U>(response);
  }

  const response = await client.patch(url, { json, ...options });
  return parseResponse<U>(response);
};

export const remove = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);
  const response = await client.delete(url, options);
  return parseResponse<T>(response);
};
