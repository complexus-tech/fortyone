import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import type { Workspace } from "@/types";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

const getCurrentWorkspace = (workspaces: Workspace[]) => {
  if (typeof window === "undefined") return null;
  const slug = window.location.hostname.split(".")[0];
  return workspaces.find((workspace) => workspace.slug === slug);
};

export const useWorkspaces = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getWorkspaces(session!.token),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
export const useCurrentWorkspace = () => {
  const { data: workspaces = [] } = useWorkspaces();
  const workspace = getCurrentWorkspace(workspaces);
  return { workspace };
};
