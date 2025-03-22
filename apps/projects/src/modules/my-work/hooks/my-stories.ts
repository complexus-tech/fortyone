import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { getMyStories } from "../queries/get-stories";

export const useMyStories = () => {
  return useQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(),
  });
};
