import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { objectiveKeys } from "../../modules/objectives/constants";
import { getObjectiveStatuses } from "../../modules/objectives/queries/statuses";

export const useObjectiveStatuses = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: objectiveKeys.statuses(workspaceSlug),
    queryFn: () => getObjectiveStatuses({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
