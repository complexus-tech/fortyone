import { useQuery } from "@tanstack/react-query";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/get-workspaces";

export const useWorkspaces = () => {
  return useQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: getWorkspaces,
  });
};
