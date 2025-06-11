import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { statusKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStatuses } from "../queries/states/get-states";
import { getTeamStatuses } from "../queries/states/get-team-states";

export const useStatuses = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: statusKeys.lists(),
    queryFn: () => getStatuses(session!),
  });
};

export const useTeamStatuses = (teamId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: statusKeys.team(teamId),
    queryFn: () => getTeamStatuses(teamId, session!),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
