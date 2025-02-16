import type { Workspace } from "@/types";

export const getCurrentWorkspace = (workspaces: Workspace[]) => {
  if (typeof window === "undefined") return null;
  const slug = window.location.hostname.split(".")[0];
  return workspaces.find((workspace) => workspace.slug === slug);
};
