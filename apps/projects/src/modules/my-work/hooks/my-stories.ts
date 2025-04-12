import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getMyStories } from "../queries/get-stories";

export const useMyStories = () => {
  return useQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
    refetchOnMount: true,
  });
};
