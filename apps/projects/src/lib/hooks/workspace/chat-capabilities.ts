"use client";

import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import {
  CHAT_CAPABILITY_RECHECK_INTERVAL_MS,
  chatCapabilitiesQueryOptions,
} from "@/lib/queries/workspaces/get-chat-capabilities";

/** Chat and attachment availability are authorized by the server for this viewer. */
export const useChatCapabilities = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const userId = session?.user.id;
  const query = useQuery({
    ...chatCapabilitiesQueryOptions({ session, workspaceSlug }, userId),
    refetchInterval: CHAT_CAPABILITY_RECHECK_INTERVAL_MS,
    refetchIntervalInBackground: false,
    refetchOnWindowFocus: "always",
    refetchOnReconnect: "always",
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
