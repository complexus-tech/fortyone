"use server";

import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { requestError } from "../fetch-error";

const apiURL = getApiUrl();

export async function inviteMembers(
  emails: string[],
  teamIds: string[],
  workspaceSlug: string,
) {
  try {
    const session = await auth();
    const cookieHeader = await getCookieHeader();
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });

    await ky.post(`${apiURL}/workspaces/${workspaceSlug}/invitations`, {
      json: {
        invitations: emails.map((email) => ({
          email,
          role: "admin",
          teamIds,
        })),
      },
      credentials: "include",
      headers,
    });

    return {
      data: null,
      error: {
        message: null,
      },
    };
  } catch (error) {
    const result = await requestError(error);
    return result;
  }
}
