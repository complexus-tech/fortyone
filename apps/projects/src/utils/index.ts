import { ApiError } from "@/lib/http/error";
import type { ApiResponse } from "@/types";

export const slugify = (text = "") => {
  return text
    .toString()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim()
    .replace(/\s+/g, "-")
    .replace(/&/g, "-and-")
    .replace(/[^\w-]+/g, "")
    .replace(/--+/g, "-");
};

export const getApiError = (error: unknown): ApiResponse<null> => {
  if (error instanceof ApiError) {
    return error.data as ApiResponse<null>;
  }
  return {
    data: null,
    error: {
      message: "An error occurred",
    },
  };
};
