import { useQuery } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import type { Team } from "../types";
import { getTeams } from "../queries/get-teams";

export const useTeams = () => {
  return useQuery<Team[]>({
    queryKey: teamKeys.lists(),
    queryFn: getTeams,
  });
};
