"use server";

import ky from "ky";
import { auth, refreshSession } from "@/auth";
import type { ApiResponse } from "@/types";
import { requestError } from "../fetch-error";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export async function acceptInvitation(token: string) {
  const session = await auth();

  try {
    const response = await ky.post(`${apiUrl}/invitations/${token}/accept`, {
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    });

    const data = await response.json<ApiResponse<null>>();
    await refreshSession();
    return data;
  } catch (error) {
    return requestError<null>(error);
  }
}
