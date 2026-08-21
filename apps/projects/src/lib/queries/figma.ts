import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  FigmaIntegration,
  StoryFigmaLink,
} from "@/modules/settings/workspace/integrations/figma/types";

export const getFigmaIntegration = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<FigmaIntegration>>(
    "integrations/figma",
    ctx,
  );
  return response.data!;
};

export const getStoryFigmaLinks = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<StoryFigmaLink[]>>(
    `stories/${storyId}/figma-links`,
    ctx,
  );
  return response.data ?? [];
};
