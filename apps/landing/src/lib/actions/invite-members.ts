"use server";

import ky from "ky";
import { auth } from "@/auth";
import { requestError } from "../fetch-error";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export async function inviteMembers(
  emails: string[],
  teamIds: string[],
  workspaceId: string,
) {
  try {
    const session = await auth();
    await ky.post(`${apiURL}/workspaces/${workspaceId}/invitations`, {
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
