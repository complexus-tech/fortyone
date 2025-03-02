"use server";

import { revalidateTag } from "next/cache";
import { linkTags } from "@/constants/keys";
import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteLinkAction = async (linkId: string, storyId: string) => {
  try {
    await remove(`links/${linkId}`);
    revalidateTag(linkTags.story(storyId));
    return linkId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete link");
  }
};
