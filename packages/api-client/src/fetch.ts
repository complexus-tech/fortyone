import ky, { type Options } from "ky";
import { ApiError } from "./error";

type ApiClient = {
  delete: (url: string, options?: Options) => Promise<Response>;
  get: (url: string, options?: Options) => Promise<Response>;
  patch: (url: string, options?: Options) => Promise<Response>;
  post: (url: string, options?: Options) => Promise<Response>;
  put: (url: string, options?: Options) => Promise<Response>;
};

const isServer = typeof window === "undefined";

const parseErrorData = async (response: Response): Promise<unknown> => {
  try {
    return await response.clone().json();
  } catch {
    try {
      return await response.text();
    } catch {
      return null;
    }
  }
};

const extractErrorMessage = (data: unknown, fallback: string): string => {
  if (data && typeof data === "object") {
    const errorData = data as {
      error?: { message?: string };
      message?: string;
    };
    if (errorData.error?.message) {
      return errorData.error.message;
    }
    if (errorData.message) {
      return errorData.message;
    }
  }
  return fallback;
};

const createKyClient = (prefixUrl?: string) => {
  const resolvedPrefixUrl = prefixUrl ?? process.env.NEXT_PUBLIC_API_URL;

  if (!resolvedPrefixUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is not configured");
  }

  return ky.create({
    prefixUrl: resolvedPrefixUrl,
    credentials: "include",
    hooks: {
      beforeRequest: isServer
        ? [
            async (request) => {
              const { headers } = await import("next/headers");
              const cookieHeader = (await headers()).get("cookie");

              if (cookieHeader && !request.headers.has("cookie")) {
                request.headers.set("cookie", cookieHeader);
              }
            },
          ]
        : [],
      beforeError: [
        async (error) => {
          const { response } = error;
          if (!response) {
            throw error;
          }

          const data = await parseErrorData(response);
          const message = extractErrorMessage(
            data,
            error.message || "Request failed",
          );

          throw new ApiError(message, response.status, data);
        },
      ],
    },
  });
};

const parseSuccessfulResponse = async <T>(
  response: Response,
): Promise<T | null> => {
  if (response.status === 204) {
    return null;
  }

  return response.json();
};

export type RequestOptions = Options;

export const createApiClient = (prefixUrl?: string): ApiClient => {
  const client = createKyClient(prefixUrl);

  return {
    delete: (url, options) => client.delete(url, options),
    get: (url, options) => client.get(url, options),
    patch: (url, options) => client.patch(url, options),
    post: (url, options) => client.post(url, options),
    put: (url, options) => client.put(url, options),
  };
};

const apiClient = createKyClient();

export const get = <T>(url: string, options?: Options) =>
  apiClient.get(url, options).json<T>();

export const post = async <T>(
  url: string,
  payload: unknown,
  options?: Options,
): Promise<T> => {
  const response =
    payload instanceof FormData
      ? await apiClient.post(url, { body: payload, ...options })
      : await apiClient.post(url, { json: payload, ...options });

  return (await parseSuccessfulResponse<T>(response)) as T;
};

export const put = <T>(url: string, payload: unknown, options?: Options) => {
  if (payload instanceof FormData) {
    return apiClient.put(url, { body: payload, ...options }).json<T>();
  }
  return apiClient.put(url, { json: payload, ...options }).json<T>();
};

export const patch = <T>(url: string, payload: unknown, options?: Options) => {
  if (payload instanceof FormData) {
    return apiClient.patch(url, { body: payload, ...options }).json<T>();
  }
  return apiClient.patch(url, { json: payload, ...options }).json<T>();
};

export const remove = async <T = unknown>(
  url: string,
  options?: Options,
): Promise<T | null> => {
  const response = await apiClient.delete(url, options);
  return parseSuccessfulResponse<T>(response);
};
