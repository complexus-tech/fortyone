import { useQuery } from "@tanstack/react-query";
import { getActivities } from "../queries/activities/get-activities";

export const useActivities = () => {
  return useQuery({
    queryKey: ["activities"],
    queryFn: getActivities,
  });
};
