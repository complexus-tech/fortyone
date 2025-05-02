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

export const createLinkAction = async (payload: NewLink) => {
  try {
    const session = await auth();
    const link = await post<NewLink, ApiResponse<Link>>(
      "links",
      payload,
      session!,
    );
    return link;
  } catch (error) {
    return getApiError(error);
  }
};
