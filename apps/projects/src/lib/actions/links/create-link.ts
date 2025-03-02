"use server";

import { revalidateTag } from "next/cache";
import { linkTags } from "@/constants/keys";
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
    revalidateTag(linkTags.story(payload.storyId));
    return link.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create link");
  }
};
