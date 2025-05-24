import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useTeamStories = (teamId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.team(teamId),
    queryFn: () => getStories(session!, { teamId }),
    enabled: Boolean(teamId),
  });
};
