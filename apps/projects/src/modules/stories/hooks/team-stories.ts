import { useSuspenseQuery } from "@tanstack/react-query";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useTeamStories = (teamId: string) => {
  return useSuspenseQuery({
    queryKey: storyKeys.team(teamId),
    queryFn: () => getStories({ teamId }),
  });
};
