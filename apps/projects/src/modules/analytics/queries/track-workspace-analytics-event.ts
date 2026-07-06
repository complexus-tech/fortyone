import { post, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  WorkspaceAnalyticsEventPayload,
  WorkspaceAnalyticsEventResponse,
} from "../types";

export const trackWorkspaceAnalyticsEvent = async (
  ctx: WorkspaceCtx,
  payload: WorkspaceAnalyticsEventPayload,
) => {
  const response = await post<
    WorkspaceAnalyticsEventPayload,
    ApiResponse<WorkspaceAnalyticsEventResponse>
  >("analytics/events", payload, ctx);

  return response.data!;
};
