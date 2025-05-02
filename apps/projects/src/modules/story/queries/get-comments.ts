import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { Comment, ApiResponse } from "@/types";

export const getStoryComments = async (id: string, session: Session) => {
  const story = await get<ApiResponse<Comment[]>>(
    `stories/${id}/comments`,
    session,
  );
  return story.data!;
};
