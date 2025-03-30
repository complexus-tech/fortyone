import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStory } from "../queries/get-story";

export const useStoryById = (id: string) => {
  return useQuery({
    queryKey: storyKeys.detail(id),
    queryFn: () => getStory(id),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
  });
};
