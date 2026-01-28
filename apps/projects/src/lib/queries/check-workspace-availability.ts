"use server";

import type { ApiResponse } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
import { requestError } from "../fetch-error";

const apiUrl = getApiUrl();

export async function checkWorkspaceAvailability(slug: string) {
  const session = await auth();
  try {
    const availability = await ky
      .get(`${apiUrl}/workspaces/check-availability?slug=${slug}`, {
        headers: {
          Authorization: `Bearer ${session?.token}`,
        },
      })
      .json<ApiResponse<{ available: boolean; slug: string }>>();

    return availability;
  } catch (error) {
    const data = await requestError<{ available: boolean; slug: string }>(
      error,
    );
    return data;
  }
}
