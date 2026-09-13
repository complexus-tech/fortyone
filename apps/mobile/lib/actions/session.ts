import type { ApiResponse, User } from "@/types";
import { get, remove } from "@/lib/http";
import { getApiURL } from "@/lib/http/config";

export const getCurrentUser = async () => {
  const response = await get<ApiResponse<User>>("auth/me", {
    useWorkspace: false,
    handleUnauthorized: false,
  });
  if (!response.data?.id)
    throw new Error("The session response did not include an account.");
  return response.data;
};
export const clearSession = async (
  sessionCookie: string,
  apiOrigin: string,
) => {
  if (getApiURL().origin !== apiOrigin) {
    throw new Error("The session belongs to a different API environment.");
  }
  await remove("users/session", {
    useWorkspace: false,
    sessionCookie,
    handleUnauthorized: false,
  });
};
