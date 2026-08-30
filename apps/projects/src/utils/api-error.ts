import { ApiError } from "api-client";
import type { ApiResponse } from "@/types/api-response";
import { reportApiErrorOutcome } from "./api-error-outcome";

export const getApiError = (error: unknown): ApiResponse<null> => {
  if (error instanceof ApiError) {
    reportApiErrorOutcome({
      certainty:
        error.status >= 400 && error.status < 500 ? "definite" : "uncertain",
      status: error.status,
    });
    return error.data as ApiResponse<null>;
  }

  reportApiErrorOutcome({ certainty: "uncertain" });
  return {
    data: null,
    error: {
      message: "An error occurred",
    },
  };
};
