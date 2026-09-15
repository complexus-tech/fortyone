import ky, { HTTPError } from "ky";
import { getApiUrl } from "@/lib/api-url";
import { requestError } from "@/lib/fetch-error";
import type { ApiResponse } from "@/types";

export class InvitationUnavailableError extends Error {}

export async function acceptUserInvitation(invitationId: string) {
  let result: ApiResponse<null>;
  try {
    const response = await ky.post(
      `${getApiUrl()}/users/me/invitations/${encodeURIComponent(invitationId)}/accept`,
      { credentials: "include" },
    );
    result = await response.json<ApiResponse<null>>();
  } catch (error) {
    if (error instanceof HTTPError && error.response.status === 404) {
      throw new InvitationUnavailableError(
        "This invitation is no longer available and has been removed from your invitations.",
      );
    }
    const failure = await requestError<null>(error);
    throw new Error(failure.error?.message || "Failed to accept invitation");
  }
  if (result.error) {
    throw new Error(result.error.message);
  }
  return result;
}
