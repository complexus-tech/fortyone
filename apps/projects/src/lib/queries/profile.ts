import ky from "ky";
import type { AuthHeaderOptions } from "@/lib/http/auth-headers";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getApiUrl } from "@/lib/api-url";
import type { ApiResponse, User } from "@/types";

const apiURL = getApiUrl();

export async function getProfile({
  token,
  cookieHeader,
}: AuthHeaderOptions = {}) {
  const headers = buildAuthHeaders({ token, cookieHeader });
  const res = await ky.get(`${apiURL}/users/profile`, {
    credentials: "include",
    headers,
  });
  const data = await res.json<ApiResponse<User>>();
  return data.data!;
}
