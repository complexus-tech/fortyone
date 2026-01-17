import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getTotalStories = async (ctx: WorkspaceCtx) => {
  try {
    const stories = await get<ApiResponse<{ count: number }>>(
      "stories/count",
      ctx,
    );
    return stories.data!.count;
  } catch {
    return 0;
  }
};
