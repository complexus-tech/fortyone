import { HTTPError, TimeoutError } from "ky";
import type { ApiResponse } from "@/types";

export const requestError = async <T>(error: unknown) => {
  if (error instanceof HTTPError) {
    let errorData: ApiResponse<T> | null = null;

    try {
      const parsed = await error.response.clone().json();
      errorData = parsed as ApiResponse<T>;
    } catch {
      errorData = null;
    }

    return {
      data: null,
      error: {
        message:
          errorData?.error?.message || `HTTP error ${error.response.status}`,
      },
    } as ApiResponse<T>;
  } else if (error instanceof TimeoutError) {
    return {
      data: null,
      error: {
        message: "Request timed out",
      },
    } as ApiResponse<T>;
  }
  // Handle network errors or other unexpected errors
  return {
    data: null,
    error: {
      message: "Failed to connect to the server",
    },
  } as ApiResponse<T>;
};
