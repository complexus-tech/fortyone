import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaceSettings } from "../../queries/workspaces/get-settings";

export const useWorkspaceSettings = () => {
  return useQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getWorkspaceSettings(),
  });
};
