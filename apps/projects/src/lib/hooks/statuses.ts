import { useQuery } from "@tanstack/react-query";
import { statusKeys } from "@/constants/keys";
import { getStatuses } from "../queries/states/get-states";

export const useStatuses = () => {
  return useQuery({
    queryKey: statusKeys.lists(),
    queryFn: getStatuses,
  });
};
