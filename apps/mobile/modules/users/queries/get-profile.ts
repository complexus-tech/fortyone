import { get } from "@/lib/http";
import type { ApiResponse, User } from "@/types";

export const getProfile = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<User>>("users/profile", {
    useWorkspace: false,
    signal,
  });
  return response.data!;
};
