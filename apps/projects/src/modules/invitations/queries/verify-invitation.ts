import "server-only";

import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { requestError } from "@/lib/fetch-error";
import type { ApiResponse } from "@/types";
import type { Invitation } from "../types";

const apiUrl = getApiUrl();

export async function verifyInvitation(token: string) {
  try {
    const response = await ky.get(`${apiUrl}/invitations/${token}`);
    return await response.json<ApiResponse<Invitation>>();
  } catch (error) {
    const response = await requestError<Invitation>(error);

    return {
      data: null,
      error: {
        message: response.error?.message || "Failed to verify invitation",
      },
    } satisfies ApiResponse<Invitation>;
  }
}
