import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useTeamMembers = (teamId?: string, search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: [
      ...memberKeys.team(workspaceSlug, teamId ?? ""),
      normalizedSearch,
    ],
    queryFn: () =>
      getTeamMembers(
        teamId!,
        { session: session!, workspaceSlug },
        normalizedSearch,
      ),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
