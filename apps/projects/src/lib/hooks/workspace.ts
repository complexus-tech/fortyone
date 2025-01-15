import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspace } from "../queries/workspaces/get-workspace";

export const useWorkspace = () => {
  return useQuery({
    queryKey: workspaceKeys.detail(),
    queryFn: () => getWorkspace(),
  });
};
