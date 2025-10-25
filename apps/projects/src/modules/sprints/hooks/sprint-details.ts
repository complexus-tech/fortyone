import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { getSprint } from "../queries/get-sprint-details";
import { useSprints } from "./sprints";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { Sprint } from "../types";

export const useSprint = (sprintId: string) => {
  const { data: session } = useSession();
  const { data: sprints = [], isPending: isSprintsPending } = useSprints();

  const existingSprint = sprints.find((sprint) => sprint.id === sprintId);

  const query = useQuery({
    queryKey: sprintKeys.detail(sprintId),
    queryFn: () => getSprint(sprintId, session!),
    enabled: !existingSprint && !isSprintsPending && Boolean(sprintId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });

  return {
    ...query,
    data: existingSprint || (query.data as Sprint), // Return existing sprint if found
    isPending: isSprintsPending || (!existingSprint && query.isPending),
  };
};
