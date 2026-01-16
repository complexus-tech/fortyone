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
    // In a path-based approach, we might need a better way to get the slug on the server
    // For now, let's try to extract it from the referer or a custom header if we can
    // Or, we can rely on the caller passing it, but let's try to keep it autonomous
    const referer = headersList.get("referer");
    if (referer) {
      const url = new URL(referer);
      subdomain = url.pathname.split("/")[1] || "";
    }
  } else {
    subdomain = window.location.pathname.split("/")[1];
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
