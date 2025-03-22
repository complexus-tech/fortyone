import { useQuery } from "@tanstack/react-query";
import { labelKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLabels } from "../queries/labels/get-labels";

export const useLabels = () => {
  return useQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
