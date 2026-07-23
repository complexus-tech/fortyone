import { headers } from "next/headers";
import { notFound } from "next/navigation";
import { getApiUrl } from "@/lib/api-url";
import {
  toPublicContributor,
  toPublicContributorCommentsPage,
  toPublicPortal,
  type ApiContributor,
  type ApiContributorCommentsPage,
  type ApiPortal,
} from "./data";
import type {
  PublicContributor,
  PublicContributorCommentsPage,
  PublicFeedbackListStatus,
  PublicPortal,
  PublicPortalWorkspace,
} from "./types";

type ApiResponse<T> = {
  data: T;
};

export type PublicPortalQuery = {
  authorId?: string;
  page?: number;
  pageSize?: number;
  search?: string;
  status?: PublicFeedbackListStatus;
  boardId?: string;
  sort?: "top" | "newest" | "oldest";
};

export type PublicPortalCachePolicy = {
  revalidateSeconds?: number;
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

const getPublicFeedbackPath = async (portalSlug: string) => {
  const workspaceSlug = await getWorkspaceSlugFromHost();

  return workspaceSlug
    ? `/workspaces/${workspaceSlug}/portals/${portalSlug}/feedback`
    : `/portals/${portalSlug}/feedback`;
};

const buildQuery = (query: PublicPortalQuery) => {
  const params = new URLSearchParams();
  if (query.authorId) params.set("authorId", query.authorId);
  if (query.page) params.set("page", String(query.page));
  if (query.pageSize) params.set("pageSize", String(query.pageSize));
  if (query.search?.trim()) params.set("search", query.search.trim());
  if (query.status) params.set("status", query.status);
  if (query.boardId) params.set("boardId", query.boardId);
  if (query.sort) params.set("sort", query.sort);

  const value = params.toString();
  return value ? `?${value}` : "";
};

const getPublicFetchOptions = ({
  revalidateSeconds,
}: PublicPortalCachePolicy) =>
  revalidateSeconds === undefined
    ? ({ cache: "no-store" } as const)
    : ({ next: { revalidate: revalidateSeconds } } as const);

export const getPublicPortal = async (
  portalSlug: string,
  query: PublicPortalQuery = {},
  cachePolicy: PublicPortalCachePolicy = {},
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
  const fetchOptions = getPublicFetchOptions(cachePolicy);

  const [workspaceResponse, feedbackResponse] = await Promise.all([
    fetch(`${apiUrl}${workspacePath}`, fetchOptions),
    fetch(`${apiUrl}${feedbackPath}`, fetchOptions),
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

export const getPublicContributor = async (
  portalSlug: string,
  authorId: string,
): Promise<PublicContributor> => {
  const apiUrl = getApiUrl();

  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to load a contributor");
  }

  const feedbackPath = await getPublicFeedbackPath(portalSlug);
  const response = await fetch(
    `${apiUrl}${feedbackPath}/contributors/${authorId}`,
    { cache: "no-store" },
  );

  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to load public feedback contributor",
      response.status,
    );
  }

  const payload = (await response.json()) as ApiResponse<ApiContributor>;
  return toPublicContributor(payload.data);
};

export const getPublicContributorComments = async (
  portalSlug: string,
  authorId: string,
  page = 1,
  pageSize = 20,
): Promise<PublicContributorCommentsPage> => {
  const apiUrl = getApiUrl();

  if (!apiUrl) {
    throw new Error(
      "NEXT_PUBLIC_API_URL is required to load contributor comments",
    );
  }

  const feedbackPath = await getPublicFeedbackPath(portalSlug);
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  const response = await fetch(
    `${apiUrl}${feedbackPath}/contributors/${authorId}/comments?${params.toString()}`,
    { cache: "no-store" },
  );

  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to load public feedback contributor comments",
      response.status,
    );
  }

  const payload =
    (await response.json()) as ApiResponse<ApiContributorCommentsPage>;
  return toPublicContributorCommentsPage(payload.data);
};

export const getPublicPortalOrNotFound = async (
  portalSlug: string,
  query: PublicPortalQuery = {},
  cachePolicy: PublicPortalCachePolicy = {},
) => {
  try {
    return await getPublicPortal(portalSlug, query, cachePolicy);
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      notFound();
    }

    throw error;
  }
};

export const getPublicContributorOrNotFound = async (
  portalSlug: string,
  authorId: string,
) => {
  try {
    return await getPublicContributor(portalSlug, authorId);
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      notFound();
    }

    throw error;
  }
};
