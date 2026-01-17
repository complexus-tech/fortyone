import { useCurrentWorkspace } from "@/lib/hooks/workspaces";

const EXTERNAL_LINK_PATTERN = /^(https?:|mailto:|tel:)/i;

export const useWorkspacePath = () => {
  const { workspace } = useCurrentWorkspace();
  const workspaceSlug = workspace?.slug ?? "";

  const withWorkspace = (path: string) => {
    if (!workspaceSlug || !path) {
      return path;
    }

    if (EXTERNAL_LINK_PATTERN.test(path)) {
      return path;
    }

    const normalizedPath = path.startsWith("/") ? path : `/${path}`;

    if (
      normalizedPath === `/${workspaceSlug}` ||
      normalizedPath.startsWith(`/${workspaceSlug}/`)
    ) {
      return normalizedPath;
    }

    return `/${workspaceSlug}${normalizedPath}`;
  };

  return { withWorkspace, workspaceSlug };
};
