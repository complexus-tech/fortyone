"use server";

import type { ApiResponse } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
import { requestError } from "../fetch-error";

const apiUrl = getApiUrl();

export async function acceptInvitation(token: string) {
  const session = await auth();

  try {
    const response = await ky.post(`${apiUrl}/invitations/${token}/accept`, {
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    });

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return requestError<null>(error);
  }
}
