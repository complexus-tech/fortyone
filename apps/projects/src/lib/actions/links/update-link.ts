"use server";

import { linkTags } from "@/constants/keys";
import { put } from "@/lib/http";
import { ApiResponse, Link } from "@/types";
import { revalidateTag } from "next/cache";

type Payload = {
  url?: string;
  title?: string;
};

export const updateLinkAction = async (
  linkId: string,
  payload: Payload,
  storyId: string,
) => {
  const _ = await put<Payload, ApiResponse<Link>>(`links/${linkId}`, payload);
  revalidateTag(linkTags.story(storyId));
  return linkId;
};
