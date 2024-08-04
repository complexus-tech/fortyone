import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { getStates } from "../queries/states/get-states";

export const useStatuses = () => {
  return useQuery({
    queryKey: storyKeys.mine(),
    queryFn: getStates,
  });
};
