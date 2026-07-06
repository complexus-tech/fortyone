import { headers } from "next/headers";
import { notFound } from "next/navigation";
import { getApiUrl } from "@/lib/api-url";
import { toPublicPortal, type ApiPortal } from "./data";
import type {
  PublicPortal,
  PublicPortalWorkspace,
  PublicRequestStatus,
} from "./types";

type ApiResponse<T> = {
  data: T;
};

export type PublicPortalQuery = {
  page?: number;
  pageSize?: number;
  search?: string;
  status?: PublicRequestStatus;
  boardId?: string;
  sort?: "top" | "newest" | "oldest";
};

const DOMAIN_SUFFIX = ".fortyone.app";

export class PublicPortalRequestError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "PublicPortalRequestError";
    this.status = status;
  }
}

export const isPublicPortalNotFoundError = (
  error: unknown,
): error is PublicPortalRequestError =>
  error instanceof PublicPortalRequestError && error.status === 404;

const getWorkspaceSlugFromHost = async () => {
  const headerList = await headers();
  const host = headerList.get("host")?.split(":")[0] ?? "";

  if (!host.endsWith(DOMAIN_SUFFIX)) return null;

  const subdomain = host.replace(DOMAIN_SUFFIX, "");
  return subdomain && subdomain !== "cloud" ? subdomain : null;
};

const buildQuery = (query: PublicPortalQuery) => {
  const params = new URLSearchParams();
  if (query.page) params.set("page", String(query.page));
  if (query.pageSize) params.set("pageSize", String(query.pageSize));
  if (query.search?.trim()) params.set("search", query.search.trim());
  if (query.status) params.set("status", query.status);
  if (query.boardId) params.set("boardId", query.boardId);
  if (query.sort) params.set("sort", query.sort);

  const value = params.toString();
  return value ? `?${value}` : "";
};

export const getPublicPortal = async (
  portalSlug: string,
  query: PublicPortalQuery = {},
): Promise<PublicPortal> => {
  const apiUrl = getApiUrl();

  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to load public feedback");
  }

  const workspaceSlug = await getWorkspaceSlugFromHost();
  const workspacePath = workspaceSlug
    ? `/workspaces/${workspaceSlug}/portal`
    : `/portals/${portalSlug}`;
  const feedbackPath = workspaceSlug
    ? `/workspaces/${workspaceSlug}/portals/${portalSlug}/feedback${buildQuery(query)}`
    : `/portals/${portalSlug}/feedback${buildQuery(query)}`;

  const [workspaceResponse, feedbackResponse] = await Promise.all([
    fetch(`${apiUrl}${workspacePath}`, { cache: "no-store" }),
    fetch(`${apiUrl}${feedbackPath}`, { cache: "no-store" }),
  ]);

  if (!feedbackResponse.ok) {
    throw new PublicPortalRequestError(
      "Failed to load public feedback portal",
      feedbackResponse.status,
    );
  }

  const workspacePayload = workspaceResponse.ok
    ? ((await workspaceResponse.json()) as ApiResponse<PublicPortalWorkspace>)
    : null;
  const feedbackPayload =
    (await feedbackResponse.json()) as ApiResponse<ApiPortal>;

  return toPublicPortal(feedbackPayload.data, workspacePayload?.data);
};

export const getPublicPortalOrNotFound = async (
  portalSlug: string,
  query: PublicPortalQuery = {},
) => {
  try {
    return await getPublicPortal(portalSlug, query);
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      notFound();
    }

    throw error;
  }
};
