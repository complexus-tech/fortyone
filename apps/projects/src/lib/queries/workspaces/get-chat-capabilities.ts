import { queryOptions } from "@tanstack/react-query";
import type { WorkspaceCtx } from "@/lib/http";
import { get } from "@/lib/http";
import { chatCapabilityKeys } from "@/constants/keys";
import type { ApiResponse } from "@/types/api-response";

export type ChatCapabilities = {
  enabled: boolean;
  attachmentsEnabled: boolean;
};

export const CHAT_CAPABILITY_RECHECK_INTERVAL_MS = 30_000;

/** Share the authorized bootstrap result with navigation, Chat, and composers. */
export const chatCapabilitiesQueryOptions = (
  ctx: WorkspaceCtx,
  userId: string | undefined,
) =>
  queryOptions({
    queryKey: chatCapabilityKeys.viewer(ctx.workspaceSlug, userId),
    queryFn: async ({ signal }) => {
      const response = await get<ApiResponse<ChatCapabilities>>(
        "chat/capabilities",
        ctx,
        { signal, cache: "no-store", retry: 0 },
      );
      if (typeof response.data?.enabled !== "boolean") {
        throw new Error("Chat availability could not be checked.");
      }
      return response.data;
    },
    enabled: Boolean(userId && ctx.workspaceSlug),
    staleTime: CHAT_CAPABILITY_RECHECK_INTERVAL_MS,
    gcTime: 5 * 60_000,
    retry: false,
  });
