"use server";

import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
import { requestError } from "../fetch-error";

const apiURL = getApiUrl();

export async function inviteMembers(
  emails: string[],
  teamIds: string[],
  workspaceSlug: string,
) {
  try {
    const session = await auth();

    await ky.post(`${apiURL}/workspaces/${workspaceSlug}/invitations`, {
      json: {
        invitations: emails.map((email) => ({
          email,
          role: "admin",
          teamIds,
        })),
      },
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
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
