"use server";

import { linkTags } from "@/constants/keys";
import { put } from "@/lib/http";
import { ApiResponse, Link } from "@/types";
import { revalidateTag } from "next/cache";

export type UpdateLink = {
  url?: string;
  title?: string;
};

export const updateLinkAction = async (
  linkId: string,
  payload: UpdateLink,
  storyId: string,
) => {
  const _ = await put<UpdateLink, ApiResponse<Link>>(
    `links/${linkId}`,
    payload,
  );
  revalidateTag(linkTags.story(storyId));
  return linkId;
};
