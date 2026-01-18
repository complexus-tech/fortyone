import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { getTeamSettings } from "../queries/get-team-settings";

export const useTeamSettings = (teamId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: teamKeys.settings(workspaceSlug, teamId),
    queryFn: () => getTeamSettings(teamId, { session: session!, workspaceSlug }),
  });
};
