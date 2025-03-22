import { useQuery } from "@tanstack/react-query";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getSprints } from "@/modules/sprints/queries/get-sprints";

export const useSprints = () => {
  return useQuery({
    queryKey: sprintKeys.lists(),
    queryFn: () => getSprints(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
