import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { teamKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getTeamSettings } from "../queries/get-team-settings";

export const useTeamSettings = (teamId: string) => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: teamKeys.settings(teamId),
    queryFn: () => getTeamSettings(teamId, session!),
    enabled: Boolean(teamId && session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
};
