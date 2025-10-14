import { useQuery } from "@tanstack/react-query";
import { getLabels } from "../queries/get-labels";
import { labelKeys } from "@/constants/keys";
import type { Label } from "@/types";

export const useLabels = () => {
  return useQuery<Label[]>({
    queryKey: labelKeys.lists(),
    queryFn: getLabels,
  });
};

export const useTeamLabels = (teamId: string) => {
  return useQuery<Label[]>({
    queryKey: labelKeys.team(teamId),
    queryFn: () => getLabels({ teamId }),
  });
};
