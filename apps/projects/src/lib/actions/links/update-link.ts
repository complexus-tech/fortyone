"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateLink = {
  url?: string;
  title?: string;
};

export const updateLinkAction = async (linkId: string, payload: UpdateLink) => {
  try {
    const session = await auth();
    const res = await put<UpdateLink, ApiResponse<Link>>(
      `links/${linkId}`,
      payload,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
