import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { requestError } from "@/lib/fetch-error";
import type { ApiResponse } from "@/types";

const apiUrl = getApiUrl();

export async function acceptInvitation(inviteToken: string) {
  try {
    const response = await ky.post(
      `${apiUrl}/invitations/${inviteToken}/accept`,
      {
        credentials: "include",
      },
    );

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return requestError<null>(error);
  }
}
