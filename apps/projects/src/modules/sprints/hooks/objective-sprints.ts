import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { getObjectiveSprints } from "../queries/get-objective-sprints";

export const useObjectiveSprints = (objectiveId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: sprintKeys.objective(objectiveId),
    queryFn: () => getObjectiveSprints(objectiveId, session!),
    enabled: Boolean(objectiveId),
  });
};
