import type { ApiResponse } from "@/types";
import { get, post, put, remove } from "@/lib/http";

export function responseData<T>(response: ApiResponse<T> | undefined): T {
  if (response?.error?.message) throw new Error(response.error.message);
  if (response?.data == null)
    throw new Error("The server returned no data. Please try again.");
  return response.data;
}
export async function readEntity<T>(path: string, signal?: AbortSignal) {
  return responseData(await get<ApiResponse<T>>(path, { signal }));
}
export async function writeEntity<T = void>(
  method: "put" | "post" | "delete",
  path: string,
  body?: unknown,
): Promise<T | undefined> {
  const response =
    method === "delete"
      ? await remove<ApiResponse<T>>(path)
      : method === "put"
        ? await put<unknown, ApiResponse<T>>(path, body)
        : await post<unknown, ApiResponse<T>>(path, body);
  if (response?.error?.message) throw new Error(response.error.message);
  return response?.data ?? undefined;
}
