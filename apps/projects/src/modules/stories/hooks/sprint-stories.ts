import { useSuspenseQuery } from "@tanstack/react-query";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useSprintStories = (sprintId: string) => {
  return useSuspenseQuery({
    queryKey: storyKeys.sprint(sprintId),
    queryFn: () => getStories({ sprintId }),
  });
};
