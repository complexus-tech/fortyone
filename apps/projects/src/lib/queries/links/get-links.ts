"use server";
import { get } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";

export const getLinks = async (storyId: string) => {
  const links = await get<ApiResponse<Link[]>>(`stories/${storyId}/links`);
  return links.data!;
};
