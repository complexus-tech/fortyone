import ky from "ky";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { getApiUrl } from "@/lib/api-url";
import { requestError } from "@/lib/fetch-error";
import type { ApiResponse } from "@/types";

const apiUrl = getApiUrl();

export async function acceptInvitation(inviteToken: string) {
  const cookieHeader = await getCookieHeader();
  try {
    const headers = buildAuthHeaders({ cookieHeader });
    const response = await ky.post(
      `${apiUrl}/invitations/${inviteToken}/accept`,
      {
        credentials: "include",
        headers,
      },
    );

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return requestError<null>(error);
  }
}
