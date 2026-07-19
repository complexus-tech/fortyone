import { getLoginUrl } from "@/utils/callback-url";
import type { PublicPortal, PublicRequest } from "./types";

const isWorkspaceSubdomainDeployment =
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

export const getBoard = (portal: PublicPortal, boardId: string) =>
  portal.boards.find((board) => board.id === boardId);

export const getRequestPathBySlug = (
  portal: PublicPortal,
  requestSlug: string,
) =>
  isWorkspaceSubdomainDeployment
    ? `/feedback/${requestSlug}`
    : `/portal/${portal.slug}/feedback/${requestSlug}`;

export const getRequestPath = (portal: PublicPortal, request: PublicRequest) =>
  getRequestPathBySlug(portal, request.slug);

export const getPortalPath = (
  portal: PublicPortal,
  path: "" | "account" | "feedback" | "roadmap" | "updates",
) => getPortalPathBySlug(portal.slug, path);

export const getPortalPathBySlug = (
  portalSlug: string,
  path: "" | "account" | "feedback" | "roadmap" | "updates",
) => {
  if (isWorkspaceSubdomainDeployment) {
    return `/${path || "feedback"}`;
  }
  return `/portal/${portalSlug}${path ? `/${path}` : ""}`;
};

export const getPortalCallbackUrl = (
  portal: PublicPortal,
  path: "account" | "feedback" | "roadmap" | "updates",
) => {
  const portalPath = getPortalPath(portal, path);

  if (!isWorkspaceSubdomainDeployment) return portalPath;

  return `https://${portal.workspace.slug}.fortyone.app${portalPath}`;
};

export const getPortalLoginUrl = (
  portal: PublicPortal,
  path: "account" | "feedback" | "roadmap" | "updates",
) => getLoginUrl(getPortalCallbackUrl(portal, path));

export const getRequestCallbackUrl = (
  portal: PublicPortal,
  request: PublicRequest,
) => {
  const requestPath = getRequestPath(portal, request);

  if (!isWorkspaceSubdomainDeployment) return requestPath;

  return `https://${portal.workspace.slug}.fortyone.app${requestPath}`;
};

export const getRequestLoginUrl = (
  portal: PublicPortal,
  request: PublicRequest,
) => getLoginUrl(getRequestCallbackUrl(portal, request));
