import ky from "ky";
import type { AuthHeaderOptions } from "@/lib/http/auth-headers";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getApiUrl } from "@/lib/api-url";
import type { ApiResponse } from "@/types";
import { requestError } from "../fetch-error";

const apiURL = getApiUrl();

export async function getAuthCode({
  token,
  cookieHeader,
}: AuthHeaderOptions = {}) {
  try {
    const headers = buildAuthHeaders({ token, cookieHeader });
    const res = await ky.get(`${apiURL}/users/session/code`, {
      credentials: "include",
      headers,
    });
    const data = await res.json<ApiResponse<{ code: string; email: string }>>();
    return data;
  } catch (error) {
    const res = await requestError<{ code: string; email: string }>(error);
    return res;
  }
}
