import { useSession } from "next-auth/react";
import { useQuery } from "@tanstack/react-query";
import type { Workspace } from "@/types";
import { DURATION_FROM_MILLISECONDS } from "@/utils";
import { getWorkspaces } from "../queries/get-workspaces";

export const workspaceKeys = {
  all: ["workspaces"] as const,
  lists: () => [...workspaceKeys.all, "list"] as const,
  settings: () => [...workspaceKeys.all, "settings"] as const,
};

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
    enabled: Boolean(session),
  });
};
export const useCurrentWorkspace = () => {
  const { data: workspaces = [] } = useWorkspaces();
  const workspace = getCurrentWorkspace(workspaces);
  return { workspace };
};
