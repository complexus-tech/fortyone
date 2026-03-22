import type { ApiResponse } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { getApiError } from "@/utils";

const apiURL = getApiUrl();

export async function linkGitHubUserAction(code: string) {
  try {
    const session = await auth();
    const cookieHeader = await getCookieHeader();
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    await ky.post(`${apiURL}/user/integrations/github/link`, {
      json: { code },
      credentials: "include",
      headers,
    });
    return { data: null } as ApiResponse<null>;
  } catch (error) {
    return getApiError(error);
  }
}

export async function unlinkGitHubUserAction() {
  try {
    const session = await auth();
    const cookieHeader = await getCookieHeader();
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    await ky.delete(`${apiURL}/user/integrations/github/link`, {
      credentials: "include",
      headers,
    });
    return { data: null } as ApiResponse<null>;
  } catch (error) {
    return getApiError(error);
  }
}
