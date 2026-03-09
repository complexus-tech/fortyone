import { get } from "api-client";
import type { AuthHeaderOptions } from "@/lib/http/auth-headers";
import type { ApiResponse, User } from "@/types";

export async function getProfile({
  token: _token,
  cookieHeader: _cookieHeader,
}: AuthHeaderOptions = {}) {
  const res = await get<ApiResponse<User>>("users/profile");
  return res.data!;
}
