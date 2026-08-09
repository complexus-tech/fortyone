import { useQuery } from "@tanstack/react-query";
import { integrationRequestKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { shouldPollRequestThread } from "../delivery-status";
import { getIntegrationRequestThread } from "../queries/get-request-thread";

const THREAD_REFRESH_INTERVAL_MS = 5_000;
const PENDING_DELIVERY_REFRESH_INTERVAL_MS = 2_000;

export const useIntegrationRequestThread = (
  requestId: string,
  enabled = true,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: integrationRequestKeys.thread(workspaceSlug, requestId),
    queryFn: () =>
      getIntegrationRequestThread(requestId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(requestId && session && enabled),
    refetchInterval: (query) =>
      shouldPollRequestThread(query.state.data)
        ? PENDING_DELIVERY_REFRESH_INTERVAL_MS
        : THREAD_REFRESH_INTERVAL_MS,
  });
};
