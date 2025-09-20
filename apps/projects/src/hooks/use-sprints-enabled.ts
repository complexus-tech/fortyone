import { useTeams } from "@/modules/teams/hooks/teams";

export const useSprintsEnabled = (teamId: string) => {
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId);
  return team?.sprintsEnabled ?? false;
};
