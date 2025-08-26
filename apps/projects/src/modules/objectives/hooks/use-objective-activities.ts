import { useInfiniteQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { objectiveKeys } from "@/modules/objectives/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectiveActivities } from "@/modules/objectives/queries/get-objective-activities";

export const useObjectiveActivitiesInfinite = (objectiveId: string) => {
  const { data: session } = useSession();

  return useInfiniteQuery({
    queryKey: objectiveKeys.activitiesInfinite(objectiveId),
    queryFn: ({ pageParam }) =>
      getObjectiveActivities(objectiveId, session!, pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.page + 1 : undefined,
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
