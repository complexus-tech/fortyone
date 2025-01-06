import { useQuery } from "@tanstack/react-query";
import { getLabels } from "../queries/labels/get-labels";
import { labelKeys } from "@/constants/keys";

export const useLabels = () => {
  return useQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(),
  });
};
