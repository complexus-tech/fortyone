import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { statusKeys } from "@/constants/keys";
import { getStatuses } from "../queries/states/get-states";

export const useStatuses = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: statusKeys.lists(),
    queryFn: () => getStatuses(session!),
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
