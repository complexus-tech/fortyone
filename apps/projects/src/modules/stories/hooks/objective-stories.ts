import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useObjectiveStories = (objectiveId: string) => {
  return useQuery({
    queryKey: storyKeys.objective(objectiveId),
    queryFn: () => getStories({ objectiveId }),
  });
};
