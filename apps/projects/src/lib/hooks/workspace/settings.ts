import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceSettings } from "../../queries/workspaces/get-settings";

export const useWorkspaceSettings = () => {
  return useQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getWorkspaceSettings(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
