import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useSprintStories = (sprintId: string) => {
  return useQuery({
    queryKey: storyKeys.sprint(sprintId),
    queryFn: () => getStories({ sprintId }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
  });
};
