import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useTeamMembers = (teamId?: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: memberKeys.team(teamId ?? ""),
    queryFn: () => getTeamMembers(teamId!, session!),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
