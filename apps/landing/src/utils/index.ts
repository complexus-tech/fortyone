import type { Session } from "next-auth";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const getRedirectUrl = (session: Session) => {
  if (session.workspaces.length === 0) {
    return "/onboarding/create";
  }
  const activeWorkspace = session.activeWorkspace || session.workspaces[0];
  if (domain.includes("localhost")) {
    return `http://${activeWorkspace.slug}.localhost:3000/my-work`;
  }
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};
