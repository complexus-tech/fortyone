"use server";

import { linkTags } from "@/constants/keys";
import { post } from "@/lib/http";
import { ApiResponse, Link } from "@/types";
import { revalidateTag } from "next/cache";

export type NewLink = {
  url: string;
  title?: string;
  storyId: string;
};

export const createLinkAction = async (payload: NewLink) => {
  const link = await post<NewLink, ApiResponse<Link>>("links", payload);
  revalidateTag(linkTags.story(payload.storyId));
  return link.data!;
};
