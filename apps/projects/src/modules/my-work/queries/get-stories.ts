import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { Story } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getMyStories = async (session: Session) => {
  const stories = await get<ApiResponse<Story[]>>("my-stories", session);
  return stories.data!;
};
