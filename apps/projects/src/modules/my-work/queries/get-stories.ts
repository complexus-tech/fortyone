import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { Story } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getMyStories = async (ctx: WorkspaceCtx) => {
  const stories = await get<ApiResponse<Story[]>>("my-stories", ctx);
  return stories.data!;
};
