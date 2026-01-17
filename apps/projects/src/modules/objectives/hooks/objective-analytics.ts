import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "../constants";
import { getObjectiveAnalytics } from "../queries/get-objective-analytics";

export const useObjectiveAnalytics = (objectiveId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: objectiveKeys.analytics(workspaceSlug, objectiveId),
    queryFn: () => getObjectiveAnalytics(objectiveId, { session: session!, workspaceSlug }),
  });
};
