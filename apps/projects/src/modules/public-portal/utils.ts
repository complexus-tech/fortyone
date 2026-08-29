import { getLoginUrl } from "@/utils/callback-url";
import type { PublicPortal, PublicRequest } from "./types";
import { NEW_FEEDBACK_QUERY_PARAM } from "./query-params";

const isWorkspaceSubdomainDeployment =
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";
const DEFAULT_AUTH_HOST = "cloud.fortyone.app";
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

export const getCrossPortalRequestHref = (
  workspaceSlug: string,
  portalSlug: string,
  requestSlug: string,
) =>
  isWorkspaceSubdomainDeployment
    ? `https://${workspaceSlug}.fortyone.app/feedback/${requestSlug}`
    : `/portal/${portalSlug}/feedback/${requestSlug}`;

export const getAuthorPathByPortalSlug = (
  portalSlug: string,
  authorId: string | null,
) => {
  if (!authorId || authorId === NIL_AUTHOR_ID) return null;

  return isWorkspaceSubdomainDeployment
    ? `/people/${authorId}`
    : `/portal/${portalSlug}/people/${authorId}`;
};

export const getAuthorPath = (portal: PublicPortal, authorId: string | null) =>
  getAuthorPathByPortalSlug(portal.slug, authorId);

export const getGlobalProfileHref = () => {
  if (!isWorkspaceSubdomainDeployment) return "/profile";

  const authHost = process.env.NEXT_PUBLIC_AUTH_HOST ?? DEFAULT_AUTH_HOST;
  return `https://${authHost}/profile`;
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

export const getPortalCallbackUrl = (
  portal: PublicPortal,
  path: "account" | "feedback" | "roadmap" | "updates",
) => {
  const portalPath = getPortalPath(portal, path);

  if (!isWorkspaceSubdomainDeployment) return portalPath;

  return `https://${portal.workspace.slug}.fortyone.app${portalPath}`;
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

export const getUpdatePathBySlug = (portalSlug: string, updateSlug: string) => {
  const updatePath = `/updates/${encodeURIComponent(updateSlug)}`;
  return isWorkspaceSubdomainDeployment
    ? updatePath
    : `/portal/${portalSlug}${updatePath}`;
};

export const getFeedbackPreferencesPath = (portalSlug: string) =>
  isWorkspaceSubdomainDeployment
    ? "/feedback/preferences"
    : `/portal/${portalSlug}/feedback/preferences`;
