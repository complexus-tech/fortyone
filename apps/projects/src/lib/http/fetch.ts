import type { Options } from "ky";
import ky from "ky";
import type { Session } from "next-auth";
import { ApiError } from "./error";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export type Ctx = {
  session: Session;
  workspaceSlug: string;
};

const createClient = (ctx: Ctx) => {
  const prefixUrl = `${apiURL}/workspaces/${ctx.workspaceSlug}/`;

  const client = ky.create({
    prefixUrl,
    headers: {
      Authorization: `Bearer ${ctx.session.token}`,
    },
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
  ctx: Ctx,
  options?: Options,
) => {
  const client = createClient(ctx);
  return client.get(url, options).json<T>();
};

export const post = async <T, U>(
  url: string,
  json: T,
  ctx: Ctx,
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
  ctx: Ctx,
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
  ctx: Ctx,
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
  ctx: Ctx,
  options?: Options,
) => {
  const client = createClient(ctx);
  return client.delete(url, options).json<T>();
};
