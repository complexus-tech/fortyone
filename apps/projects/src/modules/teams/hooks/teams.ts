import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { Team } from "../types";
import { getTeams } from "../queries/get-teams";
import { getPublicTeams } from "../queries/get-public-teams";

export const useTeams = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery<Team[]>({
    queryKey: teamKeys.lists(workspaceSlug),
    queryFn: () => getTeams({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const usePublicTeams = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery<Team[]>({
    queryKey: teamKeys.public(workspaceSlug),
    queryFn: () => getPublicTeams({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
