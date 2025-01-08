import { storyKeys } from "@/modules/stories/constants";
import { useQuery } from "@tanstack/react-query";
import { getStoryComments } from "../queries/get-comments";

export const useStoryComments = (id: string) => {
  return useQuery({
    queryKey: storyKeys.comments(id),
    queryFn: () => getStoryComments(id),
  });
};
