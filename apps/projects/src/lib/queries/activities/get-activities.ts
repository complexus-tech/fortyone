import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getActivities = async (ctx: WorkspaceCtx) => {
  const activities = await get<ApiResponse<StoryActivity[]>>("activities", ctx);
  return activities.data!;
};
