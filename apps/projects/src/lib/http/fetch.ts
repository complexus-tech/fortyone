"use server";
import type { Options } from "ky";
import ky from "ky";
import { headers } from "next/headers";
import { cache } from "react";
import { auth } from "@/auth";
import { ApiError } from "./error";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

// Cache client creation with workspace context using React's cache function
const getClientContext = cache(async () => {
  const headersList = await headers();
  const host = headersList.get("host");
  const subdomain = host ? host.split(".")[0] : "";
  const session = await auth();

  const workspaces = session?.workspaces || [];
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
  );

  const prefixUrl = `${apiURL}/workspaces/${workspace?.id}/`;

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

  const authHeaders = session
    ? {
        Authorization: `Bearer ${session.token}`,
      }
    : {};

  return { client, authHeaders };
});

export const get = async <T>(url: string, options?: Options) => {
  const { client, authHeaders } = await getClientContext();
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };
  return client.get(url, mergedOptions).json<T>();
};

export const post = async <T, U>(url: string, json: T, options?: Options) => {
  const { client, authHeaders } = await getClientContext();
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

export const put = async <T, U>(url: string, json: T, options?: Options) => {
  const { client, authHeaders } = await getClientContext();
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

export const patch = async <T, U>(url: string, json: T, options?: Options) => {
  const { client, authHeaders } = await getClientContext();
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

export const remove = async <T>(url: string, options?: Options) => {
  const { client, authHeaders } = await getClientContext();
  const mergedOptions = {
    headers: {
      ...authHeaders,
      ...(options?.headers || {}),
    },
    ...options,
  };
  return client.delete(url, mergedOptions).json<T>();
};
