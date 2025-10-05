import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { useAuthStore } from "@/store/auth";

export const useWorkspaces = () => {
  return useQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: getWorkspaces,
  });
};

export const useCurrentWorkspace = () => {
  const { data: workspaces = [] } = useWorkspaces();
  const workspaceSlug = useAuthStore((state) => state.workspace);

  const currentWorkspace = workspaces.find((w) => w.slug === workspaceSlug);

  return { workspace: currentWorkspace };
};
