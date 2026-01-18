import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import { getSprintAnalytics } from "../queries/get-sprint-analytics";

export const useSprintAnalytics = (sprintId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: sprintKeys.analytics(workspaceSlug, sprintId),
    queryFn: () => getSprintAnalytics(sprintId, { session: session!, workspaceSlug }),
  });
};
