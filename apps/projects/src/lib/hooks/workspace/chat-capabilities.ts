"use client";

import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { get } from "@/lib/http";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { chatCapabilityKeys } from "@/constants/keys";
import type { ApiResponse } from "@/types/api-response";

export type ChatCapabilities = {
  enabled: boolean;
  attachmentsEnabled: boolean;
};
const RECHECK_INTERVAL_MS = 30_000;

/** Chat and attachment availability are authorized by the server for this viewer. */
export const useChatCapabilities = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const userId = session?.user.id;
  const query = useQuery({
    queryKey: chatCapabilityKeys.viewer(workspaceSlug, userId),
    queryFn: async ({ signal }) => {
      const response = await get<ApiResponse<ChatCapabilities>>(
        "chat/capabilities",
        { session, workspaceSlug },
        { signal, cache: "no-store", retry: 0 },
      );
      if (typeof response.data?.enabled !== "boolean") {
        throw new Error("Chat availability could not be checked.");
      }
      return response.data;
    },
    enabled: Boolean(userId && workspaceSlug),
    staleTime: RECHECK_INTERVAL_MS,
    gcTime: RECHECK_INTERVAL_MS,
    refetchInterval: RECHECK_INTERVAL_MS,
    refetchIntervalInBackground: false,
    refetchOnWindowFocus: "always",
    refetchOnReconnect: "always",
    retry: false,
  });

  return {
    ...query,
    // A failed recheck must not keep an old enabled result usable.
    enabled: Boolean(userId && query.isSuccess && query.data.enabled),
    attachmentsEnabled: Boolean(
      userId &&
        query.isSuccess &&
        query.data.enabled &&
        query.data.attachmentsEnabled,
    ),
  };
};
