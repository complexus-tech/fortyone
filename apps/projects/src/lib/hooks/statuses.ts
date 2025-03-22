import { useQuery } from "@tanstack/react-query";
import { statusKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStatuses } from "../queries/states/get-states";

export const useStatuses = () => {
  return useQuery({
    queryKey: statusKeys.lists(),
    queryFn: () => getStatuses(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamStatuses = (teamId: string) => {
  const { data: statuses = [] } = useStatuses();
  return useQuery({
    queryKey: statusKeys.team(teamId),
    queryFn: () => statuses.filter((status) => status.teamId === teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    enabled: Boolean(teamId),
  });
};
