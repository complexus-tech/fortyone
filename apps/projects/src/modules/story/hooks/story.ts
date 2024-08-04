import { storyKeys } from "@/modules/stories/constants";
import { useQuery } from "@tanstack/react-query";
import { getStory } from "../queries/get-story";

export const useStoryById = (id: string) => {
  return useQuery({
    queryKey: storyKeys.detail(id),
    queryFn: () => getStory(id),
  });
};
