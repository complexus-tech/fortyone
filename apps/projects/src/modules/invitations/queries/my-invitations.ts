import ky from "ky";
import type { AuthHeaderOptions } from "@/lib/http/auth-headers";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getApiUrl } from "@/lib/api-url";
import type { ApiResponse } from "@/types";
import type { Invitation } from "../types";

const apiURL = getApiUrl();

export async function getMyInvitations({
  token,
  cookieHeader,
}: AuthHeaderOptions = {}) {
  try {
    const headers = buildAuthHeaders({ token, cookieHeader });
    const response = await ky.get(`${apiURL}/users/me/invitations`, {
      credentials: "include",
      headers,
    });
    const invitations = await response.json<ApiResponse<Invitation[]>>();
    return invitations.data ?? [];
  } catch (error) {
    return [];
  }
}
