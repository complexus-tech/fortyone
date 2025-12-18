import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DetailedStory } from "../types";

export const getStory = async (id: string, session: Session) => {
  try {
    const story = await get<ApiResponse<DetailedStory>>(
      `stories/${id}`,
      session,
    );
    return story.data;
  } catch (error) {
    return null;
  }
};

export const getStoryRef = async (ref: string, session: Session) => {
  try {
    const story = await get<ApiResponse<DetailedStory>>(
      `story-by-ref/${ref}`,
      session,
    );
    return story.data;
  } catch (error) {
    return null;
  }
};
