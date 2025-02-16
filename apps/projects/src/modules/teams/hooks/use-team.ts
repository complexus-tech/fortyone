import { useQuery } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import type { Team } from "../types";
import { useTeams } from "./teams";

export const useTeam = (teamId: string) => {
  const { data: teams = [] } = useTeams();

  return useQuery({
    queryKey: teamKeys.detail(teamId),
    queryFn: () => teams.find((team: Team) => team.id === teamId),
    enabled: Boolean(teamId),
  });
};
