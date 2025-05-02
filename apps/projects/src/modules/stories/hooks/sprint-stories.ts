import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useSprintStories = (sprintId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.sprint(sprintId),
    queryFn: () => getStories(session!, { sprintId }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
    refetchOnMount: true,
  });
};
