import { headers } from "next/headers";
import { notFound } from "next/navigation";
import { getApiUrl } from "@/lib/api-url";
import {
  toPublicContributor,
  toPublicContributorCommentsPage,
  toPublicPortal,
  toPublicPortalUpdate,
  type ApiContributor,
  type ApiContributorCommentsPage,
  type ApiFeedbackUpdate,
  type ApiPortal,
} from "./data";
import { getFeedbackSessionAuthorization } from "./guest-session";
import type {
  PublicContributor,
  PublicContributorCommentsPage,
  PublicFeedbackListStatus,
  PublicPortal,
  PublicPortalUpdate,
  PublicPortalWorkspace,
  SimilarPublicFeedback,
} from "./types";

type ApiResponse<T> = {
  data: T;
};

type ApiUpdatesPage = {
  updates: ApiFeedbackUpdate[];
  hasMore: boolean;
  unreadCount?: number;
};

export type PublicPortalUpdatesPage = {
  updates: PublicPortalUpdate[];
  hasMore: boolean;
  unreadCount: number;
};

export type PublicFeedbackCanonicalItem = {
  itemId: string;
  itemSlug: string;
  merged: boolean;
};

export type PublicPortalQuery = {
  authorId?: string;
  itemId?: string;
  page?: number;
  pageSize?: number;
  search?: string;
  status?: PublicFeedbackListStatus;
  boardId?: string;
  sort?: "top" | "newest" | "oldest";
  view?: "summary";
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
  if (query.itemId) params.set("itemId", query.itemId);
  if (query.page) params.set("page", String(query.page));
  if (query.pageSize) params.set("pageSize", String(query.pageSize));
  if (query.search?.trim()) params.set("search", query.search.trim());
  if (query.status) params.set("status", query.status);
  if (query.boardId) params.set("boardId", query.boardId);
  if (query.sort) params.set("sort", query.sort);
  if (query.view) params.set("view", query.view);

  const value = params.toString();
  return value ? `?${value}` : "";
};

const getPublicFetchOptions = ({
  revalidateSeconds,
}: PublicPortalCachePolicy) =>
  revalidateSeconds === undefined
    ? ({ cache: "no-store" } as const)
    : ({ next: { revalidate: revalidateSeconds } } as const);

const fetchPublicFeedbackPortal = async (
  apiUrl: string,
  feedbackPath: string,
  cachePolicy: PublicPortalCachePolicy,
) => {
  const response = await fetch(
    `${apiUrl}${feedbackPath}`,
    getPublicFetchOptions(cachePolicy),
  );
  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to load public feedback portal",
      response.status,
    );
  }
  const payload = (await response.json()) as ApiResponse<ApiPortal>;
  return payload.data;
};

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

  const [workspaceResponse, apiPortal] = await Promise.all([
    fetch(`${apiUrl}${workspacePath}`, fetchOptions),
    fetchPublicFeedbackPortal(apiUrl, feedbackPath, cachePolicy),
  ]);

  const workspacePayload = workspaceResponse.ok
    ? ((await workspaceResponse.json()) as ApiResponse<PublicPortalWorkspace>)
    : null;

  return toPublicPortal(apiPortal, workspacePayload?.data);
};

export const getPublicFeedbackPortal = async (
  portalSlug: string,
  query: PublicPortalQuery = {},
  cachePolicy: PublicPortalCachePolicy = {},
) => {
  const apiUrl = getApiUrl();
  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to load public feedback");
  }
  const feedbackPath = `${await getPublicFeedbackPath(portalSlug)}${buildQuery(query)}`;
  const apiPortal = await fetchPublicFeedbackPortal(
    apiUrl,
    feedbackPath,
    cachePolicy,
  );
  return toPublicPortal(apiPortal);
};

export const getSimilarPublicFeedback = async (
  portalSlug: string,
  {
    description,
    limit = 3,
    title,
  }: { description?: string; limit?: number; title: string },
): Promise<SimilarPublicFeedback[]> => {
  const apiUrl = getApiUrl();
  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to find similar feedback");
  }

  const params = new URLSearchParams({
    limit: String(limit),
    title: title.trim(),
  });
  if (description?.trim()) params.set("description", description.trim());
  const response = await fetch(
    `${apiUrl}/portals/${portalSlug}/feedback/similar?${params.toString()}`,
    { cache: "no-store" },
  );
  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to find similar feedback",
      response.status,
    );
  }

  const payload = (await response.json()) as ApiResponse<
    SimilarPublicFeedback[]
  >;
  return payload.data;
};

export const getPublicFeedbackCanonicalItem = async (
  portalSlug: string,
  itemReference: string,
): Promise<PublicFeedbackCanonicalItem> => {
  const apiUrl = getApiUrl();
  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to resolve feedback");
  }

  const response = await fetch(
    `${apiUrl}/portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(itemReference)}/canonical`,
    { cache: "no-store" },
  );
  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to resolve public feedback",
      response.status,
    );
  }

  const payload =
    (await response.json()) as ApiResponse<PublicFeedbackCanonicalItem>;
  return payload.data;
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

export const getPublicPortalUpdates = async (
  portalSlug: string,
  page = 1,
  pageSize = 20,
): Promise<PublicPortalUpdatesPage> => {
  const apiUrl = getApiUrl();
  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to load public updates");
  }

  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  const authorization = await getFeedbackSessionAuthorization(portalSlug);
  const response = await fetch(
    `${apiUrl}/portals/${encodeURIComponent(portalSlug)}/feedback/updates?${params.toString()}`,
    {
      cache: "no-store",
      ...(authorization ? { headers: { Authorization: authorization } } : {}),
    },
  );
  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to load public feedback updates",
      response.status,
    );
  }

  const payload = (await response.json()) as ApiResponse<ApiUpdatesPage>;
  return {
    updates: payload.data.updates.map(toPublicPortalUpdate),
    hasMore: payload.data.hasMore,
    unreadCount: payload.data.unreadCount ?? 0,
  };
};

export const getPublicPortalUpdate = async (
  portalSlug: string,
  updateSlug: string,
): Promise<PublicPortalUpdate> => {
  const apiUrl = getApiUrl();
  if (!apiUrl) {
    throw new Error("NEXT_PUBLIC_API_URL is required to load a public update");
  }

  const response = await fetch(
    `${apiUrl}/portals/${encodeURIComponent(portalSlug)}/feedback/updates/${encodeURIComponent(updateSlug)}`,
    { cache: "no-store" },
  );
  if (!response.ok) {
    throw new PublicPortalRequestError(
      "Failed to load public feedback update",
      response.status,
    );
  }
  const payload = (await response.json()) as ApiResponse<ApiFeedbackUpdate>;
  return toPublicPortalUpdate(payload.data);
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

export const getPublicFeedbackPortalOrNotFound = async (
  portalSlug: string,
  query: PublicPortalQuery = {},
  cachePolicy: PublicPortalCachePolicy = {},
) => {
  try {
    return await getPublicFeedbackPortal(portalSlug, query, cachePolicy);
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      notFound();
    }

    throw error;
  }
};

export const getPublicFeedbackCanonicalItemOrNotFound = async (
  portalSlug: string,
  itemReference: string,
) => {
  try {
    return await getPublicFeedbackCanonicalItem(portalSlug, itemReference);
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

export const getPublicPortalUpdateOrNotFound = async (
  portalSlug: string,
  updateSlug: string,
) => {
  try {
    return await getPublicPortalUpdate(portalSlug, updateSlug);
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      notFound();
    }

    throw error;
  }
};
