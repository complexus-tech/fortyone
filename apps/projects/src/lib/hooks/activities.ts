import { useSuspenseQuery } from "@tanstack/react-query";
import { getActivities } from "../queries/activities/get-activities";

export const useActivities = () => {
  return useSuspenseQuery({
    queryKey: ["activities"],
    queryFn: () => getActivities(),
  });
};
