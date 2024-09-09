import { useQuery } from "@tanstack/react-query";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";

export const useObjectives = () => {
  return useQuery({
    queryKey: ["objectives"],
    queryFn: getObjectives,
  });
};
