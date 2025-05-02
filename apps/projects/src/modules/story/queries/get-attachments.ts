import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryAttachment } from "../types";

export const getStoryAttachments = async (id: string, session: Session) => {
  const story = await get<ApiResponse<StoryAttachment[]>>(
    `stories/${id}/attachments`,
    session,
  );
  return story.data!;
};
