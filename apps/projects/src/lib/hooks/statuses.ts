import { useQuery } from "@tanstack/react-query";
import { statusKeys } from "@/constants/keys";
import { getStatuses } from "../queries/states/get-states";

export const useStatuses = () => {
  return useQuery({
    queryKey: statusKeys.lists(),
    queryFn: getStatuses,
  });
};

export const useTeamStatuses = (teamId: string) => {
  const { data: statuses = [] } = useStatuses();
  return useQuery({
    queryKey: statusKeys.team(teamId),
    queryFn: () => statuses.filter((status) => status.teamId === teamId),
    enabled: Boolean(teamId),
  });
};
