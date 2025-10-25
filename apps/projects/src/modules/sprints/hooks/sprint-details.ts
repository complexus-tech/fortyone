import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { getSprint } from "../queries/get-sprint-details";
import { useSprints } from "./sprints";
import { useTeamSprints } from "./team-sprints";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { Sprint } from "../types";

export const useSprint = (sprintId: string | null, teamId?: string) => {
  const { data: session } = useSession();
  const { data: allSprints = [], isPending: isAllSprintsPending } =
    useSprints();
  const { data: teamSprints = [], isPending: isTeamSprintsPending } =
    useTeamSprints(teamId || "");

  // Use team sprints if teamId provided, otherwise use all sprints
  const sprints = teamId ? teamSprints : allSprints;
  const isSprintsPending = teamId ? isTeamSprintsPending : isAllSprintsPending;

  const existingSprint = sprints.find((sprint) => sprint.id === sprintId);

  const query = useQuery({
    queryKey: sprintKeys.detail(sprintId ?? ""),
    queryFn: () => getSprint(sprintId ?? "", session!),
    enabled: !existingSprint && !isSprintsPending && Boolean(sprintId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });

  return {
    ...query,
    data: existingSprint || (query.data as Sprint), // Return existing sprint if found
    isPending: isSprintsPending || (!existingSprint && query.isPending),
  };
};
