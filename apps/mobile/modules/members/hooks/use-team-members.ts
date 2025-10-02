import { useQuery } from "@tanstack/react-query";
import { getTeamMembers } from "../queries/get-members";
import { memberKeys } from "@/constants/keys";

export const useTeamMembers = (teamId?: string) => {
  return useQuery({
    queryKey: memberKeys.team(teamId ?? ""),
    queryFn: () => getTeamMembers(teamId!),
    enabled: Boolean(teamId),
  });
};
