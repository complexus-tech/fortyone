import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import type { Workspace } from "@/types";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

const DOMAIN_SUFFIX = ".fortyone.app";
const RESERVED_SUBDOMAINS = new Set(["cloud"]);

export const getCurrentWorkspace = (workspaces: Workspace[]) => {
  if (typeof window === "undefined") return null;
  const pathnameSegments = window.location.pathname.split("/").filter(Boolean);
  const host = window.location.host.split(":")[0];
  const slugFromPath = pathnameSegments[0];
  const slugFromSubdomain = host.endsWith(DOMAIN_SUFFIX)
    ? host.replace(DOMAIN_SUFFIX, "")
    : undefined;
  const slug =
    slugFromPath ||
    (slugFromSubdomain && !RESERVED_SUBDOMAINS.has(slugFromSubdomain)
      ? slugFromSubdomain
      : undefined);

  if (!slug) return null;

  return workspaces.find(
    (workspace) => workspace.slug.toLowerCase() === slug.toLowerCase(),
  );
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
