import { HTTPError, TimeoutError } from "ky";
import type { ApiResponse } from "@/types";

export const requestError = async (error: unknown) => {
  if (error instanceof HTTPError) {
    // Handle HTTP errors (4xx, 5xx responses)
    const errorData = await error.response.json<ApiResponse<null>>();
    return {
      data: null,
      error: {
        message:
          errorData.error?.message || `HTTP error ${error.response.status}`,
      },
    };
  } else if (error instanceof TimeoutError) {
    return {
      data: null,
      error: {
        message: "Request timed out",
      },
    };
  }
  // Handle network errors or other unexpected errors
  return {
    data: null,
    error: {
      message: "Failed to connect to the server",
    },
  };
};
