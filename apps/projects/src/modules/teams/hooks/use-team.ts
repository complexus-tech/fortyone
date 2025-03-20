import { useQuery } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { Team } from "../types";
import { useTeams } from "./teams";

export const useTeam = (teamId: string) => {
  const { data: teams = [] } = useTeams();

  return useQuery({
    queryKey: teamKeys.detail(teamId),
    queryFn: () => teams.find((team: Team) => team.id === teamId),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
};
