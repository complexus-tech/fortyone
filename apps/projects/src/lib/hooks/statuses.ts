import { useQuery } from "@tanstack/react-query";
import { getStatuses } from "../queries/states/get-states";

export const useStatuses = () => {
  return useQuery({
    queryKey: ["statuses"],
    queryFn: getStatuses,
  });
};
