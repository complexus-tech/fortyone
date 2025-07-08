import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { GroupedStoryParams } from "../types";
import { storyKeys } from "../constants";
import { getGroupedStories } from "../queries/get-grouped-stories";

export const useGroupedStories = (params: GroupedStoryParams) => {
  const { data: session } = useSession();

  const queryKey = [...storyKeys.grouped(), params] as const;

  return useQuery({
    queryKey,
    queryFn: () => getGroupedStories(session!, params),
    staleTime: 1000 * 60 * 2,
  });
};
