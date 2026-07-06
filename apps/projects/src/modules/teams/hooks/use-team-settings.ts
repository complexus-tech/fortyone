import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { getTeamSettings } from "../queries/get-team-settings";

export const useTeamSettings = (teamId?: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const settingsTeamId = teamId ?? "";

  return useQuery({
    queryKey: teamKeys.settings(workspaceSlug, settingsTeamId),
    queryFn: () =>
      getTeamSettings(settingsTeamId, { session: session!, workspaceSlug }),
    enabled: Boolean(teamId),
  });
};
