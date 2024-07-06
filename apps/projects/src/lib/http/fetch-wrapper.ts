import { HttpError } from "./error";
const apiURL = process.env.NEXT_PUBLIC_API_URL;

type ResponseType = "json" | "blob";

async function http<T>(
  path: string,
  config: RequestInit,
  responseType?: ResponseType,
  retries = 1,
): Promise<T> {
  const fullPath = apiURL + path;
  const request = new Request(fullPath, config);
  const response: Response = await fetch(fullPath, config);

  if (!response.ok) {
    if (retries > 0) {
      return await http<T>(fullPath, config, responseType, retries - 1);
    }
    const errJson = await response.json();
    const err = HttpError.fromRequest(request, response, errJson);
    throw err;
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
  body: T,
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
  body: T,
  config?: RequestInit,
  retries = 1,
): Promise<U> {
  const init = {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    ...config,
  };
  return await http<U>(path, init, "json", retries);
}
