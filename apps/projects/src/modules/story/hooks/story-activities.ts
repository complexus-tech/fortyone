import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryActivities } from "../queries/get-activities";

export const useStoryActivities = (id: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.activities(id),
    queryFn: () => getStoryActivities(id, session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
