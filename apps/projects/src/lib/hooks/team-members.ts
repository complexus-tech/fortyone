import { useQuery } from "@tanstack/react-query";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useTeamMembers = (teamId?: string) => {
  return useQuery({
    queryKey: memberKeys.team(teamId ?? ""),
    queryFn: () => getTeamMembers(teamId!),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
