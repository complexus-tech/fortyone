import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";
import { getTeamObjectives } from "../queries/get-team-objectives";

export const useObjectives = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: objectiveKeys.list(workspaceSlug),
    queryFn: () => getObjectives({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamObjectives = (teamId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: objectiveKeys.team(workspaceSlug, teamId),
    queryFn: () => getTeamObjectives(teamId, { session: session!, workspaceSlug }),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
