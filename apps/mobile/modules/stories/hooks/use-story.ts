import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getStory } from "../queries/get-story";

export const useStory = (id: string) => {
  return useQuery({
    queryKey: storyKeys.detail(id),
    queryFn: ({ signal }) => getStory(id, signal),
    enabled: Boolean(id),
  });
};
