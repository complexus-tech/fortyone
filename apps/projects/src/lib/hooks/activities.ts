import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getActivities } from "../queries/activities/get-activities";

export const useActivities = () => {
  return useQuery({
    queryKey: ["activities"],
    queryFn: () => getActivities(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
