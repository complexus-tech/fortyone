"use client";

import type { InfiniteData } from "@tanstack/react-query";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { PublicPortal, PublicPortalFilters, PublicRequest } from "./types";
import { publicPortalKeys, toPublicPortalFilterKey } from "./query-keys";

type ApiResponse<T> = {
  data: T;
};

const PAGE_SIZE = 20;
const PUBLIC_PORTAL_STALE_TIME = DURATION_FROM_MILLISECONDS.MINUTE * 5;

export type PublicFeedbackPages = InfiniteData<PublicPortal, number>;

const buildFeedbackPageUrl = ({
  filters,
  page,
  pageSize = PAGE_SIZE,
  portalSlug,
}: {
  filters: PublicPortalFilters;
  page: number;
  pageSize?: number;
  portalSlug: string;
}) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
    sort: filters.sort,
  });
  if (filters.boardId) params.set("boardId", filters.boardId);
  if (filters.search.trim()) params.set("search", filters.search.trim());
  if (filters.status) params.set("status", filters.status);

  return `/api/public-portal/${portalSlug}?${params.toString()}`;
};

export const fetchPublicFeedbackPage = async ({
  filters,
  page,
  pageSize,
  portalSlug,
}: {
  filters: PublicPortalFilters;
  page: number;
  pageSize?: number;
  portalSlug: string;
}) => {
  const response = await fetch(
    buildFeedbackPageUrl({ filters, page, pageSize, portalSlug }),
  );
  if (!response.ok) {
    throw new Error("Unable to load feedback");
  }

  const payload = (await response.json()) as ApiResponse<PublicPortal>;
  return payload.data;
};

const filtersMatch = (
  first: PublicPortalFilters,
  second: PublicPortalFilters,
) => {
  const firstKey = toPublicPortalFilterKey(first);
  const secondKey = toPublicPortalFilterKey(second);

  return (
    firstKey.boardId === secondKey.boardId &&
    firstKey.search === secondKey.search &&
    firstKey.sort === secondKey.sort &&
    firstKey.status === secondKey.status
  );
};

export const usePublicFeedbackList = ({
  filters,
  initialFilters,
  initialPortal,
}: {
  filters: PublicPortalFilters;
  initialFilters: PublicPortalFilters;
  initialPortal: PublicPortal;
}) => {
  const initialData = filtersMatch(filters, initialFilters)
    ? {
        pages: [initialPortal],
        pageParams: [1],
      }
    : undefined;

  return useInfiniteQuery({
    queryKey: publicPortalKeys.feedbackList(initialPortal.slug, filters),
    queryFn: ({ pageParam }) =>
      fetchPublicFeedbackPage({
        filters,
        page: pageParam,
        portalSlug: initialPortal.slug,
      }),
    initialData: () => initialData,
    initialPageParam: 1,
    getNextPageParam: (lastPage, _pages, lastPageParam) =>
      lastPage.requestsHasMore ? lastPageParam + 1 : undefined,
    staleTime: PUBLIC_PORTAL_STALE_TIME,
  });
};

export const usePublicFeedbackDetail = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) =>
  useQuery({
    queryKey: publicPortalKeys.feedbackDetail(portal.slug, request.id),
    queryFn: async () => {
      const nextPortal = await fetchPublicFeedbackPage({
        filters: {
          search: request.slug,
          sort: "top",
          status: request.status,
        },
        page: 1,
        pageSize: 1,
        portalSlug: portal.slug,
      });
      const nextRequest = nextPortal.requests.find(
        (item) => item.id === request.id || item.slug === request.slug,
      );

      if (!nextRequest) {
        throw new Error("Unable to load feedback");
      }
      return nextRequest;
    },
    initialData: request,
    staleTime: PUBLIC_PORTAL_STALE_TIME,
  });
