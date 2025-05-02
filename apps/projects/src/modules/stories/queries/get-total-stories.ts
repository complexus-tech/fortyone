import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getTotalStories = async (session: Session) => {
  try {
    const stories = await get<ApiResponse<{ count: number }>>(
      "stories/count",
      session,
    );
    return stories.data!.count;
  } catch {
    return 0;
  }
};
