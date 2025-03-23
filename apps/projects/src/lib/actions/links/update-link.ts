import { put } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";

export type UpdateLink = {
  url?: string;
  title?: string;
};

export const updateLinkAction = async (linkId: string, payload: UpdateLink) => {
  try {
    const res = await put<UpdateLink, ApiResponse<Link>>(
      `links/${linkId}`,
      payload,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
