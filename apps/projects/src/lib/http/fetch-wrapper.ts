import { auth } from "@/auth";
import { HttpError } from "./error";
const apiURL = process.env.NEXT_PUBLIC_API_URL;

type ResponseType = "json" | "blob";

async function http<T>(
  path: string,
  config: RequestInit,
  responseType?: ResponseType,
  retries = 1,
): Promise<T> {
  let requestConfig: RequestInit = config;

  if (process.env.NODE_ENV === "development") {
    requestConfig = {
      ...config,
      next: {
        revalidate: 1,
      },
    };
  }

  const session = await auth();
  if (session) {
    requestConfig.headers = {
      ...requestConfig.headers,
      Authorization: `Bearer ${session.token}`,
    };

    path = path.startsWith("/")
      ? `/workspaces/${session.activeWorkspace.id}` + path
      : path;
  }

  const fullPath = path.startsWith("/") ? apiURL + path : path;
  const request = new Request(fullPath, requestConfig);
  const response: Response = await fetch(fullPath, requestConfig);

  if (!response.ok) {
    if (retries > 0) {
      return await http<T>(fullPath, requestConfig, responseType, retries - 1);
    }
    const errJson = await response.json();
    const err = HttpError.fromRequest(request, response, errJson);
    throw err;
  }

  // check if it's 204, return empty object
  if (response.status === 204) {
    return {
      success: true,
    } as any;
  }

  // may error if there is no body, return empty array
  if (responseType === "blob") return (await response.blob()) as any;
  return await response.json();
}

export async function get<T>(
  path: string,
  config?: RequestInit,
  responseType?: ResponseType,
  retries = 1,
): Promise<T> {
  const init = { method: "GET", ...config };
  return await http<T>(path, init, responseType, retries);
}

type Options = {
  raw: boolean;
};
export async function post<T, U>(
  path: string,
  body: T = {} as T,
  config?: RequestInit,
  options: Options = { raw: false },
  retries = 1,
): Promise<U> {
  const init = {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: options.raw ? (body as any) : JSON.stringify(body),
    ...config,
  };
  return await http<U>(path, init, "json", retries);
}

export async function put<T, U>(
  path: string,
  body: T,
  config?: RequestInit,
  retries = 1,
): Promise<U> {
  const init = {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    ...config,
  };
  return await http<U>(path, init, "json", retries);
}

export async function patch<T, U>(
  path: string,
  body: T,
  config?: RequestInit,
  retries = 1,
): Promise<U> {
  const init = {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    ...config,
  };
  return await http<U>(path, init, "json", retries);
}

export async function remove<T, U>(
  path: string,
  config?: RequestInit,
  retries = 1,
): Promise<U> {
  const init = {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    ...config,
  };
  return await http<U>(path, init, "json", retries);
}
