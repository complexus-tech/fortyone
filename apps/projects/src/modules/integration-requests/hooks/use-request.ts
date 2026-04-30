import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import { getIntegrationRequest } from "../queries/get-request";

export const useIntegrationRequest = (requestId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: integrationRequestKeys.detail(workspaceSlug, requestId),
    queryFn: () => getIntegrationRequest(requestId, { session: session!, workspaceSlug }),
    enabled: Boolean(requestId && session),
  });
};
