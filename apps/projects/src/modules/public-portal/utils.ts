import type { PublicPortal, PublicRequest } from "./types";

export const getBoard = (portal: PublicPortal, boardId: string) =>
  portal.boards.find((board) => board.id === boardId);

export const getRequestPath = (portal: PublicPortal, request: PublicRequest) =>
  `/portal/${portal.slug}/requests/${request.slug}`;

export const getPortalPath = (
  portal: PublicPortal,
  path: "" | "requests" | "roadmap" | "updates",
) => `/portal/${portal.slug}${path ? `/${path}` : ""}`;
