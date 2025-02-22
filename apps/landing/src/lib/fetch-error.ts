import { HTTPError, TimeoutError } from "ky";
import type { ApiResponse } from "@/types";

export const requestError = async <T>(error: unknown) => {
  if (error instanceof HTTPError) {
    // Handle HTTP errors (4xx, 5xx responses)
    const errorData = await error.response.json<ApiResponse<T>>();
    return {
      data: null,
      error: {
        message:
          errorData.error?.message || `HTTP error ${error.response.status}`,
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
