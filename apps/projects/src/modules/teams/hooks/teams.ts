import { useQuery } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import type { Team } from "../types";
import { getTeams } from "../queries/get-teams";
import { getPublicTeams } from "../queries/get-public-teams";

export const useTeams = () => {
  return useQuery<Team[]>({
    queryKey: teamKeys.lists(),
    queryFn: getTeams,
  });
};

export const usePublicTeams = () => {
  return useQuery<Team[]>({
    queryKey: teamKeys.public(),
    queryFn: getPublicTeams,
  });
};
