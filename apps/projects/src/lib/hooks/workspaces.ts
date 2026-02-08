import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import type { Workspace } from "@/types";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useWorkspacePath } from "@/hooks";

export const getCurrentWorkspace = (workspaces: Workspace[], slug: string) => {
  return workspaces.find(
    (workspace) => workspace.slug.toLowerCase() === slug.toLowerCase(),
  );
};

export const useWorkspaces = () => {
  return useQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getWorkspaces(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};

export const useCurrentWorkspace = () => {
  const { data: workspaces = [] } = useWorkspaces();
  const { workspaceSlug } = useWorkspacePath();
  const workspace = getCurrentWorkspace(workspaces, workspaceSlug);
  return { workspace };
};
