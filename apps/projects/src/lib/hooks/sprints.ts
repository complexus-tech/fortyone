import { useQuery } from "@tanstack/react-query";
import { getSprints } from "@/modules/sprints/queries/get-sprints";

export const useSprints = () => {
  return useQuery({
    queryKey: ["sprints"],
    queryFn: getSprints,
  });
};
