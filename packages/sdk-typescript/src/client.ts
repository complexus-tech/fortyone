import createClient from "openapi-fetch";
import type { Middleware } from "openapi-fetch";

import { DEFAULT_BASE_URL } from "./generated/metadata.js";
import type { paths } from "./generated/schema.js";
import { createRetryingFetch, type RetryOptions } from "./retry.js";

export type FortyOneClient = ReturnType<typeof createClient<paths>>;

export interface ClientOptions {
  token: string;
  baseUrl?: string;
  fetch?: typeof globalThis.fetch;
  retry?: false | RetryOptions;
  allowInsecureLoopback?: boolean;
}

export const createFortyOneClient = (options: ClientOptions): FortyOneClient => {
  const token = validateToken(options.token);
  const baseUrl = normalizeBaseUrl(
    options.baseUrl ?? DEFAULT_BASE_URL,
    options.allowInsecureLoopback ?? false,
  );
  const fetchImplementation =
    options.retry === false
      ? (options.fetch ?? globalThis.fetch)
      : createRetryingFetch(
          options.fetch ?? globalThis.fetch,
          options.retry ?? {},
        );
  const client = createClient<paths>({ baseUrl, fetch: fetchImplementation });
  const authentication: Middleware = {
    onRequest({ request }) {
      request.headers.set("Authorization", `Bearer ${token}`);
      if (!request.headers.has("Accept")) {
        request.headers.set("Accept", "application/json");
      }
      return request;
    },
  };
  client.use(authentication);
  return client;
};

const validateToken = (value: string): string => {
  if (
    value.length === 0 ||
    value !== value.trim() ||
    /[\u0000-\u0020\u007f]/u.test(value)
  ) {
    throw new Error("FortyOne API token is missing or malformed");
  }
  return value;
};

const normalizeBaseUrl = (value: string, allowInsecureLoopback: boolean) => {
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error("FortyOne API base URL must be an absolute URL");
  }
  if (parsed.username || parsed.password || parsed.search || parsed.hash) {
    throw new Error(
      "FortyOne API base URL must not include credentials, a query, or a fragment",
    );
  }
  if (
    parsed.protocol !== "https:" &&
    !(
      allowInsecureLoopback &&
      parsed.protocol === "http:" &&
      isLoopbackHost(parsed.hostname)
    )
  ) {
    throw new Error(
      "FortyOne API base URL must use HTTPS (HTTP is available only for explicitly enabled loopback tests)",
    );
  }
  return parsed.toString().replace(/\/$/u, "");
};

const isLoopbackHost = (hostname: string) =>
  hostname === "localhost" ||
  hostname === "127.0.0.1" ||
  hostname === "[::1]";
