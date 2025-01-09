import { useQuery } from "@tanstack/react-query";
import { labelKeys } from "@/constants/keys";
import { getLabels } from "../queries/labels/get-labels";

export const useLabels = () => {
  return useQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(),
  });
};
