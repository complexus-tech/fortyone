import type { Session } from "next-auth";
import type { Invitation } from "@/types";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const getRedirectUrl = (
  session: Session,
  invitations: Invitation[] = [],
) => {
  if (session.workspaces.length === 0) {
    if (invitations.length > 0) {
      return `/onboarding/join?token=${invitations[0].token}`;
    }
    return "/onboarding/create";
  }
  const activeWorkspace = session.activeWorkspace || session.workspaces[0];
  if (domain.includes("localhost")) {
    return `http://${activeWorkspace.slug}.localhost:3000/my-work`;
  }
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};

export const buildWorkspaceUrl = (slug: string) => {
  if (domain.includes("localhost")) {
    return `http://${slug}.localhost:3000/my-work`;
  }
  return `https://${slug}.${domain}/my-work`;
};
