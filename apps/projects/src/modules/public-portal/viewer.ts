import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getRedirectUrl } from "@/utils";
import { getFeedbackSetupHref } from "./feedback-setup";
import { getPortalPathBySlug } from "./utils";
import type { PublicPortalViewer } from "./types";

export const getPublicPortalViewer = async (
  portalSlug: string,
): Promise<PublicPortalViewer | null> => {
  const session = await auth();

  if (!session) {
    return null;
  }

  const workspaces = await getWorkspaces();
  const activeWorkspace =
    workspaces.find(
      (workspace) => workspace.id === session.user.lastUsedWorkspaceId,
    ) ?? workspaces.at(0);

  return {
    id: session.user.id,
    name: session.user.fullName || session.user.username || session.user.name,
    email: session.user.email,
    avatarUrl: session.user.image,
    appHref: activeWorkspace
      ? getRedirectUrl(workspaces, [], session.user.lastUsedWorkspaceId)
      : undefined,
    accountHref: getPortalPathBySlug(portalSlug, "account"),
    feedbackSetupHref: getFeedbackSetupHref(
      workspaces,
      session.user.lastUsedWorkspaceId,
    ),
  };
};
