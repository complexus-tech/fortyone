import { useParams } from "next/navigation";

const EXTERNAL_LINK_PATTERN = /^(https?:|mailto:|tel:)/i;

export const useWorkspacePath = () => {
  const params = useParams<{ workspaceSlug?: string }>();
  const workspaceSlug = params?.workspaceSlug?.toLowerCase() ?? "";

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
