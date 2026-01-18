"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";

export type NewLink = {
  url: string;
  title?: string;
  storyId: string;
};

export const createLinkAction = async (
  payload: NewLink,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const link = await post<NewLink, ApiResponse<Link>>("links", payload, ctx);
    return link;
  } catch (error) {
    return getApiError(error);
  }
};
