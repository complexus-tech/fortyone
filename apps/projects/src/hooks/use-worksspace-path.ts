import { useParams } from "next/navigation";

const EXTERNAL_LINK_PATTERN = /^(https?:|mailto:|tel:)/i;

export const useWorkspacePath = () => {
  const workspaceSlug = useParams<{ workspaceSlug: string }>();

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
