import { useQuery } from "@tanstack/react-query";
import { getStatuses, getTeamStatuses } from "../queries/get-statuses";
import { statusKeys } from "@/constants/keys";

export const useStatuses = () => {
  return useQuery({
    queryKey: statusKeys.lists(),
    queryFn: getStatuses,
  });
};

export const useTeamStatuses = (teamId: string) => {
  return useQuery({
    queryKey: statusKeys.team(teamId),
    queryFn: () => getTeamStatuses(teamId),
    enabled: Boolean(teamId),
  });
};
