"use server";

import { revalidateTag } from "next/cache";
import { linkTags } from "@/constants/keys";
import { remove } from "@/lib/http";

export const deleteLinkAction = async (linkId: string, storyId: string) => {
  const _ = await remove(`links/${linkId}`);
  revalidateTag(linkTags.story(storyId));
  return linkId;
};
