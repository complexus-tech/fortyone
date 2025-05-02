import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { objectiveKeys } from "../../modules/objectives/constants";
import { getObjectiveStatuses } from "../../modules/objectives/queries/statuses";

export const useObjectiveStatuses = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: () => getObjectiveStatuses(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
