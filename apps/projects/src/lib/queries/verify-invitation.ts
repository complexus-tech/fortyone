import type { ApiResponse } from "@/types";
import type { Invitation } from "@/modules/invitations/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { requestError } from "../fetch-error";

const apiUrl = getApiUrl();

export async function verifyInvitation(token: string) {
  try {
    const response = await ky.get(`${apiUrl}/invitations/${token}`);
    const data = await response.json<ApiResponse<Invitation>>();
    return data;
  } catch (error) {
    const res = await requestError<Invitation>(error);
    return {
      data: null,
      error: {
        message: res.error?.message || "Failed to verify invitation",
      },
    };
  }
}
