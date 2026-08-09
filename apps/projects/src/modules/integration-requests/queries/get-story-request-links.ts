import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { IntegrationRequestProviderThread } from "../types";

export const getStoryIntegrationRequestLinks = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<IntegrationRequestProviderThread[]>>(
    `stories/${storyId}/integration-request-links`,
    ctx,
  );
  return response.data ?? [];
};
