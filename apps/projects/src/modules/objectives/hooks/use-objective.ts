import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { objectiveKeys } from "../constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getObjective } from "../queries/get-objective";
import { useObjectives } from "./use-objectives";
import { useTeamObjectives } from "./use-objectives";

export const useObjective = (objectiveId: string | null, teamId?: string) => {
  const { data: session } = useSession();
  const { data: allObjectives = [], isPending: isAllObjectivesPending } =
    useObjectives();
  const { data: teamObjectives = [], isPending: isTeamObjectivesPending } =
    useTeamObjectives(teamId || "");

  // Use team objectives if teamId provided, otherwise use all objectives
  const objectives = teamId ? teamObjectives : allObjectives;
  const isObjectivesPending = teamId
    ? isTeamObjectivesPending
    : isAllObjectivesPending;

  const existingObjective = objectives.find(
    (objective) => objective.id === objectiveId,
  );

  const query = useQuery({
    queryKey: objectiveKeys.objective(objectiveId ?? ""),
    queryFn: () => getObjective(objectiveId ?? "", session!),
    enabled: !existingObjective && !isObjectivesPending && Boolean(objectiveId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });

  return {
    ...query,
    data: existingObjective || query.data, // Return existing objective if found
    isPending: isObjectivesPending || (!existingObjective && query.isPending),
  };
};
