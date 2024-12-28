import { useQuery } from "@tanstack/react-query";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../contants";

export const useObjectives = () => {
  return useQuery({
    queryKey: objectiveKeys.list(),
    queryFn: getObjectives,
  });
};
