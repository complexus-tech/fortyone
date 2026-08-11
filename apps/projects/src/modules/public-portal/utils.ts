import { getLoginUrl } from "@/utils/callback-url";
import type { PublicPortal, PublicRequest } from "./types";
import { NEW_FEEDBACK_QUERY_PARAM } from "./query-params";

const isWorkspaceSubdomainDeployment =
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";
const FORTYONE_DOMAIN = "fortyone.app";
const NIL_AUTHOR_ID = "00000000-0000-0000-0000-000000000000";

export const getBoard = (portal: PublicPortal, boardId: string) =>
  portal.boards.find((board) => board.id === boardId);

export const getRequestPathBySlugs = (
  portalSlug: string,
  requestSlug: string,
) =>
  isWorkspaceSubdomainDeployment
    ? `/feedback/${requestSlug}`
    : `/portal/${portalSlug}/feedback/${requestSlug}`;

export const getRequestPathBySlug = (
  portal: PublicPortal,
  requestSlug: string,
) => getRequestPathBySlugs(portal.slug, requestSlug);

export const getRequestPath = (portal: PublicPortal, request: PublicRequest) =>
  getRequestPathBySlug(portal, request.slug);

export const getAuthorPathByPortalSlug = (
  portalSlug: string,
  authorId: string,
) => {
  if (!authorId || authorId === NIL_AUTHOR_ID) return null;

  return isWorkspaceSubdomainDeployment
    ? `/people/${authorId}`
    : `/portal/${portalSlug}/people/${authorId}`;
};

export const getAuthorPath = (portal: PublicPortal, authorId: string) =>
  getAuthorPathByPortalSlug(portal.slug, authorId);

export const getViewerProfilePathByPortalSlug = (portalSlug: string) =>
  isWorkspaceSubdomainDeployment
    ? "/people/me"
    : `/portal/${portalSlug}/people/me`;

export const getViewerProfilePath = (portal: PublicPortal) =>
  getViewerProfilePathByPortalSlug(portal.slug);

export const getViewerProfileHrefByPortalSlug = (portalSlug: string) => {
  const profilePath = getViewerProfilePathByPortalSlug(portalSlug);

  if (!isWorkspaceSubdomainDeployment) return profilePath;

  return `https://${portalSlug}.${FORTYONE_DOMAIN}${profilePath}`;
};

export const getPortalPath = (
  portal: PublicPortal,
  path: "" | "account" | "feedback" | "roadmap" | "updates",
) => getPortalPathBySlug(portal.slug, path);

export const getPortalPathBySlug = (
  portalSlug: string,
  path: "" | "account" | "feedback" | "roadmap" | "updates",
) => {
  const routePath = path === "roadmap" ? "feedback/roadmap" : path;

  if (isWorkspaceSubdomainDeployment) {
    return `/${routePath || "feedback"}`;
  }
  return `/portal/${portalSlug}${routePath ? `/${routePath}` : ""}`;
};

export const getPortalAccountPathBySlug = (portalSlug: string) => {
  const accountPath = getPortalPathBySlug(portalSlug, "account");
  const params = new URLSearchParams({ portal: portalSlug });

  return `${accountPath}?${params.toString()}`;
};

export const getPortalCallbackUrl = (
  portal: PublicPortal,
  path: "account" | "feedback" | "roadmap" | "updates",
) => {
  const portalPath = getPortalPath(portal, path);

  if (!isWorkspaceSubdomainDeployment) return portalPath;

  return `https://${portal.workspace.slug}.fortyone.app${portalPath}`;
};

export const getViewerProfileCallbackUrl = (portal: PublicPortal) => {
  const profilePath = getViewerProfilePath(portal);

  if (!isWorkspaceSubdomainDeployment) return profilePath;

  return `https://${portal.workspace.slug}.fortyone.app${profilePath}`;
};

export const getNewFeedbackCallbackUrl = (portal: PublicPortal) => {
  const feedbackUrl = getPortalCallbackUrl(portal, "feedback");
  const separator = feedbackUrl.includes("?") ? "&" : "?";

  return `${feedbackUrl}${separator}${NEW_FEEDBACK_QUERY_PARAM}=true`;
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
