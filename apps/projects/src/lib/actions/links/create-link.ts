import { post } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";

export type NewLink = {
  url: string;
  title?: string;
  storyId: string;
};

export const createLinkAction = async (payload: NewLink) => {
  try {
    const link = await post<NewLink, ApiResponse<Link>>("links", payload);
    return link;
  } catch (error) {
    return getApiError(error);
  }
};
