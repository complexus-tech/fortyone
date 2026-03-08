import { get } from "api-client";
import type { AuthHeaderOptions } from "@/lib/http/auth-headers";
import type { ApiResponse } from "@/types";
import { requestError } from "../fetch-error";

export async function getAuthCode({
  token: _token,
  cookieHeader: _cookieHeader,
}: AuthHeaderOptions = {}) {
  try {
    const res =
      await get<ApiResponse<{ code: string; email: string }>>(
        "users/session/code",
      );
    return res;
  } catch (error) {
    const res = await requestError<{ code: string; email: string }>(error);
    return res;
  }
}
