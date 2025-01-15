import { useQuery } from "@tanstack/react-query";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { sprintKeys } from "@/constants/keys";

export const useSprints = () => {
  return useQuery({
    queryKey: sprintKeys.lists(),
    queryFn: getSprints,
  });
};
