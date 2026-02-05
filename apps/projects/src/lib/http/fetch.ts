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

export const get = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);
  return client.get(url, options).json<T>();
};

export const post = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);

  if (json instanceof FormData) {
    return client.post(url, { body: json, ...options }).json<U>();
  }

  return client.post(url, { json, ...options }).json<U>();
};

export const put = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);

  if (json instanceof FormData) {
    return client.put(url, { body: json, ...options }).json<U>();
  }

  return client.put(url, { json, ...options }).json<U>();
};

export const patch = async <T, U>(
  url: string,
  json: T,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);

  if (json instanceof FormData) {
    return client.patch(url, { body: json, ...options }).json<U>();
  }

  return client.patch(url, { json, ...options }).json<U>();
};

export const remove = async <T>(
  url: string,
  ctx: WorkspaceCtx,
  options?: Options,
) => {
  const client = createClient(ctx);
  return client.delete(url, options).json<T>();
};
