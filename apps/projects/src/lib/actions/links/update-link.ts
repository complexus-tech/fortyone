"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";

export type UpdateLink = {
  url?: string;
  title?: string;
};

export const updateLinkAction = async (linkId: string, payload: UpdateLink) => {
  try {
    await put<UpdateLink, ApiResponse<Link>>(`links/${linkId}`, payload);
    return linkId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update link");
  }
};
