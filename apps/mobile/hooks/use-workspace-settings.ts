import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaceSettings } from "@/lib/queries/get-workspace-settings";

export const useWorkspaceSettings = () => {
  return useQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: getWorkspaceSettings,
    staleTime: 1000 * 60 * 15, // 15 minutes
  });
};
