import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { teamKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { Team } from "../types";
import { getTeams } from "../queries/get-teams";
import { getPublicTeams } from "../queries/get-public-teams";

export const useTeams = () => {
  const { data: session } = useSession();
  return useQuery<Team[]>({
    queryKey: teamKeys.lists(),
    queryFn: () => getTeams(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const usePublicTeams = () => {
  const { data: session } = useSession();
  return useQuery<Team[]>({
    queryKey: teamKeys.public(),
    queryFn: () => getPublicTeams(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
