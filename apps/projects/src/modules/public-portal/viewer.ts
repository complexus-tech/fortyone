import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getRedirectUrl, withWorkspacePath } from "@/utils";
import type { PublicPortalViewer } from "./types";

export const getPublicPortalViewer =
  async (): Promise<PublicPortalViewer | null> => {
    const session = await auth();

    if (!session) {
      return null;
    }

    const workspaces = await getWorkspaces();
    const activeWorkspace =
      workspaces.find(
        (workspace) => workspace.id === session.user.lastUsedWorkspaceId,
      ) ?? workspaces[0];
    const workspaceSlug = activeWorkspace.slug;

    return {
      name: session.user.fullName || session.user.username || session.user.name,
      email: session.user.email,
      avatarUrl: session.user.image,
      appHref: getRedirectUrl(workspaces, [], session.user.lastUsedWorkspaceId),
      accountHref: withWorkspacePath("/settings/account", workspaceSlug),
      notificationsHref: withWorkspacePath("/notifications", workspaceSlug),
    };
  };
