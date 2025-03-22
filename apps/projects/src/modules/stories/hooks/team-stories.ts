import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useTeamStories = (teamId: string) => {
  return useQuery({
    queryKey: storyKeys.team(teamId),
    queryFn: () => getStories({ teamId }),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
