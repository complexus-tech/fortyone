"use server";

import { revalidateTag } from "next/cache";
import { linkTags } from "@/constants/keys";
import { put } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";

export type UpdateLink = {
  url?: string;
  title?: string;
};

export const updateLinkAction = async (
  linkId: string,
  payload: UpdateLink,
  storyId: string,
) => {
  try {
    await put<UpdateLink, ApiResponse<Link>>(`links/${linkId}`, payload);
    revalidateTag(linkTags.story(storyId));
    return linkId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update link");
  }
};
