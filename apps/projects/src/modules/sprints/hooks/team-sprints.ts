import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamSprints } from "../queries/get-team-sprints";

export const useTeamSprints = (teamId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: sprintKeys.team(teamId),
    queryFn: () => getTeamSprints(teamId, session!),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
