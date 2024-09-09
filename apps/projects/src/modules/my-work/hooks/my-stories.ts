import { useQuery } from "@tanstack/react-query";
import { getMyStories } from "../queries/get-stories";
import { storyKeys } from "@/modules/stories/constants";

export const useMyStories = () => {
  return useQuery({
    queryKey: storyKeys.mine(),
    queryFn: getMyStories,
  });
};
