import type { Workspace } from "@/types/workspace";
import { DEFAULT_WORKSPACE_PATH } from "@/shared/routing/workspace";
import { getSafeCallbackUrl } from "./callback-url";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

type InvitationRedirectCandidate = {
  token?: string;
};

export const getRedirectUrl = (
  workspaces: Workspace[],
  invitations: InvitationRedirectCandidate[] = [],
  lastUsedWorkspaceId?: string,
  callbackUrl?: string,
) => {
  const safeCallbackUrl = getSafeCallbackUrl(callbackUrl);
  if (safeCallbackUrl) {
    return safeCallbackUrl;
  }

  if (workspaces.length === 0) {
    if (invitations.length > 0) {
      return `/onboarding/join?token=${invitations[0].token}`;
    }
    return "/onboarding/create";
  }
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ||
    workspaces[0];

  return buildWorkspaceUrl(activeWorkspace.slug);
};

export const buildWorkspaceUrl = (
  slug: string,
  path = DEFAULT_WORKSPACE_PATH,
) => {
  if (isFortyOneApp) {
    return `https://${slug}.fortyone.app${path}`;
  }

  return `/${slug}${path}`;
};

export const withWorkspacePath = (path: string, slug?: string) => {
  if (!slug || isFortyOneApp) {
    return path;
  }

  if (path.startsWith(`/${slug}`)) {
    return path;
  }

  if (path.startsWith("/")) {
    return `/${slug}${path}`;
  }

  return `/${slug}/${path}`;
};
