import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjectives } from "../queries/get-objectives";
import { objectiveKeys } from "../constants";
import { getTeamObjectives } from "../queries/get-team-objectives";

export const useObjectives = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: objectiveKeys.list(),
    queryFn: () => getObjectives(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useTeamObjectives = (teamId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: objectiveKeys.team(teamId),
    queryFn: () => getTeamObjectives(teamId, session!),
    enabled: Boolean(teamId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
