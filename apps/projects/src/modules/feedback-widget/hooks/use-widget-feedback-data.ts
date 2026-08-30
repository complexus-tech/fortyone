"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type {
  PublicFeedbackListStatus,
  PublicPortalSort,
  PublicRequest,
} from "@/shared/feedback-widget/types";
import { getWidgetFeedbackPageAction } from "../actions";
import type {
  WidgetRoadmap,
  WidgetRoadmapPagination,
  WidgetRoadmapStatus,
} from "../components/types";
import {
  createInitialRoadmapPagination,
  INITIAL_ROADMAP_VISIBLE_COUNT,
  mergeRequests,
  replaceRequest,
} from "../components/utils";

type UseWidgetFeedbackDataOptions = {
  initialRequests: PublicRequest[];
  initialRoadmap: WidgetRoadmap;
  initialRoadmapPagination?: WidgetRoadmapPagination;
  portalSlug: string;
};

export const useWidgetFeedbackData = ({
  initialRequests,
  initialRoadmap,
  initialRoadmapPagination,
  portalSlug,
}: UseWidgetFeedbackDataOptions) => {
  const [requests, setRequests] = useState(initialRequests);
  const [homeRequests, setHomeRequests] = useState(initialRequests);
  const [search, setSearch] = useState("");
  const [feedbackSort, setFeedbackSort] = useState<PublicPortalSort>("top");
  const [feedbackStatus, setFeedbackStatus] =
    useState<PublicFeedbackListStatus>("active");
  const [isFeedbackLoading, setIsFeedbackLoading] = useState(false);
  const [feedbackError, setFeedbackError] = useState("");
  const feedbackQueryRef = useRef(0);
  const feedbackFiltersRef = useRef({
    search: "",
    sort: "top" as PublicPortalSort,
    status: "active" as PublicFeedbackListStatus,
  });
  const [roadmapItems, setRoadmapItems] = useState(initialRoadmap);
  const [roadmapPageState, setRoadmapPageState] = useState(
    initialRoadmapPagination ?? createInitialRoadmapPagination(),
  );
  const [visibleRoadmapCounts, setVisibleRoadmapCounts] = useState<
    Record<WidgetRoadmapStatus, number>
  >({
    completed: INITIAL_ROADMAP_VISIBLE_COUNT,
    in_progress: INITIAL_ROADMAP_VISIBLE_COUNT,
    planned: INITIAL_ROADMAP_VISIBLE_COUNT,
  });
  const [loadingRoadmapStatus, setLoadingRoadmapStatus] =
    useState<WidgetRoadmapStatus | null>(null);
  const [roadmapError, setRoadmapError] = useState("");

  const syncRequest = useCallback((updatedRequest: PublicRequest) => {
    setRequests((current) => replaceRequest(current, updatedRequest));
    setHomeRequests((current) => replaceRequest(current, updatedRequest));
    setRoadmapItems((current) => ({
      completed: replaceRequest(current.completed, updatedRequest),
      in_progress: replaceRequest(current.in_progress, updatedRequest),
      planned: replaceRequest(current.planned, updatedRequest),
    }));
  }, []);

  const prependRequest = useCallback((request: PublicRequest) => {
    setRequests((current) => [request, ...current]);
    setHomeRequests((current) => [request, ...current]);
  }, []);

  useEffect(() => {
    const nextFilters = {
      search,
      sort: feedbackSort,
      status: feedbackStatus,
    };
    const previousFilters = feedbackFiltersRef.current;
    if (
      previousFilters.search === nextFilters.search &&
      previousFilters.sort === nextFilters.sort &&
      previousFilters.status === nextFilters.status
    ) {
      return;
    }
    feedbackFiltersRef.current = nextFilters;
    const queryId = feedbackQueryRef.current + 1;
    feedbackQueryRef.current = queryId;
    const timeout = window.setTimeout(
      () => {
        setIsFeedbackLoading(true);
        setFeedbackError("");
        void getWidgetFeedbackPageAction({
          page: 1,
          portalSlug,
          search,
          sort: feedbackSort,
          status: feedbackStatus,
        })
          .then((response) => {
            if (feedbackQueryRef.current !== queryId) return;
            if (!response.data) {
              setFeedbackError(
                response.error?.message ?? "Unable to load feedback.",
              );
              return;
            }
            setRequests(response.data.requests);
          })
          .catch(() => {
            if (feedbackQueryRef.current === queryId) {
              setFeedbackError("Unable to load feedback.");
            }
          })
          .finally(() => {
            if (feedbackQueryRef.current === queryId) {
              setIsFeedbackLoading(false);
            }
          });
      },
      search.trim() ? 250 : 0,
    );

    return () => {
      window.clearTimeout(timeout);
    };
  }, [feedbackSort, feedbackStatus, portalSlug, search]);

  const loadMoreRoadmap = useCallback(
    async (status: WidgetRoadmapStatus) => {
      const items = roadmapItems[status];
      const visibleCount = visibleRoadmapCounts[status];
      if (visibleCount < items.length) {
        setVisibleRoadmapCounts((current) => ({
          ...current,
          [status]: Math.min(
            items.length,
            current[status] + INITIAL_ROADMAP_VISIBLE_COUNT,
          ),
        }));
        return;
      }

      const pagination = roadmapPageState[status];
      if (!pagination.hasMore || loadingRoadmapStatus) return;
      setLoadingRoadmapStatus(status);
      setRoadmapError("");
      const response = await getWidgetFeedbackPageAction({
        page: pagination.nextPage,
        portalSlug,
        search: "",
        sort: "newest",
        status,
      }).catch(() => null);
      setLoadingRoadmapStatus(null);
      if (!response?.data) {
        setRoadmapError("Unable to load more roadmap items.");
        return;
      }
      setRoadmapItems((current) => ({
        ...current,
        [status]: mergeRequests(current[status], response.data.requests),
      }));
      setRoadmapPageState((current) => ({
        ...current,
        [status]: {
          hasMore: response.data.hasMore,
          nextPage: response.data.nextPage,
        },
      }));
      setVisibleRoadmapCounts((current) => ({
        ...current,
        [status]: current[status] + INITIAL_ROADMAP_VISIBLE_COUNT,
      }));
    },
    [
      loadingRoadmapStatus,
      portalSlug,
      roadmapItems,
      roadmapPageState,
      visibleRoadmapCounts,
    ],
  );

  return {
    feedbackError,
    feedbackSort,
    feedbackStatus,
    homeRequests,
    isFeedbackLoading,
    loadMoreRoadmap,
    loadingRoadmapStatus,
    prependRequest,
    requests,
    roadmapError,
    roadmapItems,
    roadmapPageState,
    search,
    setFeedbackSort,
    setFeedbackStatus,
    setSearch,
    syncRequest,
    visibleRoadmapCounts,
  };
};
