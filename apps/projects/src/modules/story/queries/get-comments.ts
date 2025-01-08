"use server";
import { get } from "@/lib/http";
import { Comment } from "@/types";
import { ApiResponse } from "@/types";

export const getStoryComments = async (id: string) => {
  const story = await get<ApiResponse<Comment[]>>(`stories/${id}/comments`);
  return story?.data;
};
