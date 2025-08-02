import type { Options } from "ky";
import ky from "ky";
import { cache } from "react";
import type { Session } from "next-auth";
import { ApiError } from "./error";
import { getHeaders } from "./header";

const apiURL = process.env.NEXT_PUBLIC_API_URL;
const isServer = typeof window === "undefined";

// Cache client creation with workspace context using React's cache function
const getClientContext = cache(async (token: string) => {
  let subdomain = "";
  if (isServer) {
    const headersList = await getHeaders();
    subdomain = headersList.get("host")?.split(".")[0] || "";
  } else {
    subdomain = window.location.hostname.split(".")[0];
  }

  const prefixUrl = `${apiURL}/workspaces/${subdomain.toLowerCase()}/`;

  const client = ky.create({
    prefixUrl,
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

  const authHeaders = {
    Authorization: `Bearer ${token}`,
  };

  return { client, authHeaders };
});

export const get = async <T>(
  url: string,
  session: Session,
  options?: Options,
) => {
  const { client, authHeaders } = await getClientContext(session.token);
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };

  return client.get(url, mergedOptions).json<T>();
};

export const post = async <T, U>(
  url: string,
  json: T,
  session: Session,
  options?: Options,
) => {
  const { client, authHeaders } = await getClientContext(session.token);
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };

  if (json instanceof FormData) {
    return client.post(url, { body: json, ...mergedOptions }).json<U>();
  }

  return client.post(url, { json, ...mergedOptions }).json<U>();
};

export const put = async <T, U>(
  url: string,
  json: T,
  session: Session,
  options?: Options,
) => {
  const { client, authHeaders } = await getClientContext(session.token);
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };
  if (json instanceof FormData) {
    return client.put(url, { body: json, ...mergedOptions }).json<U>();
  }
  return client.put(url, { json, ...mergedOptions }).json<U>();
};

export const patch = async <T, U>(
  url: string,
  json: T,
  session: Session,
  options?: Options,
) => {
  const { client, authHeaders } = await getClientContext(session.token);
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };
  if (json instanceof FormData) {
    return client.patch(url, { body: json, ...mergedOptions }).json<U>();
  }
  return client.patch(url, { json, ...mergedOptions }).json<U>();
};

export const remove = async <T>(
  url: string,
  session: Session,
  options?: Options,
) => {
  const { client, authHeaders } = await getClientContext(session.token);
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };
  return client.delete(url, mergedOptions).json<T>();
};
