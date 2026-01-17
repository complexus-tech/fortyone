import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import { getObjectiveSprints } from "../queries/get-objective-sprints";

export const useObjectiveSprints = (objectiveId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: sprintKeys.objective(workspaceSlug, objectiveId),
    queryFn: () =>
      getObjectiveSprints(objectiveId, { session: session!, workspaceSlug }),
    enabled: Boolean(objectiveId),
  });
};
