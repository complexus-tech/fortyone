import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useTeamStories = (teamId: string) => {
  return useQuery({
    queryKey: storyKeys.team(teamId),
    queryFn: () => getStories({ teamId }),
  });
};
