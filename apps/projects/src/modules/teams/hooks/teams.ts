import { useQuery } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { Team } from "../types";
import { getTeams } from "../queries/get-teams";
import { getPublicTeams } from "../queries/get-public-teams";

export const useTeams = () => {
  return useQuery<Team[]>({
    queryKey: teamKeys.lists(),
    queryFn: () => getTeams(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const usePublicTeams = () => {
  return useQuery<Team[]>({
    queryKey: teamKeys.public(),
    queryFn: () => getPublicTeams(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
