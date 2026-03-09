import { get } from "api-client";
import type { ApiResponse, CurrentUser } from "./types";

export const fetchCurrentUser = async () => {
  const response = await get<ApiResponse<CurrentUser>>("auth/me");
  return response.data;
};

export const getCurrentUser = async () => {
  try {
    return await fetchCurrentUser();
  } catch {
    return null;
  }
};
