import ky from "ky";
import type { Session } from "next-auth";
import { getApiUrl } from "@/lib/api-url";
import type { ApiResponse } from "@/types";
import { requestError } from "../fetch-error";

const apiURL = getApiUrl();

export async function getAuthCode(session: Session) {
  try {
    const res = await ky.get(`${apiURL}/users/session/code`, {
      headers: {
        Authorization: `Bearer ${session.token}`,
      },
    });
    const data = await res.json<ApiResponse<{ code: string; email: string }>>();
    return data;
  } catch (error) {
    const res = await requestError<{ code: string; email: string }>(error);
    return res;
  }
}
