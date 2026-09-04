import type { Workspace } from "@/types/workspace";
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
    return "/account";
  }
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ||
    workspaces[0];

  if (isFortyOneApp) {
    return `https://${activeWorkspace.slug}.fortyone.app/my-work`;
  }

  return `/${activeWorkspace.slug}/my-work`;
};

export const buildWorkspaceUrl = (slug: string, path = "/my-work") => {
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
