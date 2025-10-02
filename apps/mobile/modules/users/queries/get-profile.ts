import { get } from "@/lib/http";
import type { ApiResponse, User } from "@/types";

export const getProfile = async () => {
  const response = await get<ApiResponse<User>>("users/profile", {
    useWorkspace: false,
  });
  return response.data!;
};
