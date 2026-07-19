import type { PublicPortal, PublicRequest } from "./types";

const isWorkspaceSubdomainDeployment =
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

export const getBoard = (portal: PublicPortal, boardId: string) =>
  portal.boards.find((board) => board.id === boardId);

export const getRequestPath = (portal: PublicPortal, request: PublicRequest) =>
  isWorkspaceSubdomainDeployment
    ? `/feedback/${request.slug}`
    : `/portal/${portal.slug}/feedback/${request.slug}`;

export const getPortalPath = (
  portal: PublicPortal,
  path: "" | "feedback" | "roadmap" | "updates",
) => {
  if (isWorkspaceSubdomainDeployment) {
    return `/${path || "feedback"}`;
  }
  return `/portal/${portal.slug}${path ? `/${path}` : ""}`;
};
