import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import { getTeamIntegrationRequests } from "../queries/get-team-requests";

export const useTeamIntegrationRequests = (
  teamId: string,
  status = "pending",
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: integrationRequestKeys.team(workspaceSlug, teamId, status),
    queryFn: () =>
      getTeamIntegrationRequests(teamId, { session: session!, workspaceSlug }, status),
    enabled: Boolean(teamId && session),
  });
};
