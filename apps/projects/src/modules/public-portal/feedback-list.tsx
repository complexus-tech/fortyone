"use client";

import type { ReactNode } from "react";
import { useEffect, useEffectEvent, useRef, useState } from "react";
import { ArrowUpDownIcon, RequestsIcon, SearchIcon } from "icons";
import { Box, Flex, Input, Text } from "ui";
import { cn } from "lib";
import { toast } from "sonner";
import { requestFilters, requestStatusMeta } from "./status";
import { PublicRequestCard } from "./request-card";
import type { PublicPortal, PublicPortalFilters, PublicRequest } from "./types";

type ApiResponse<T> = {
  data: T;
};

const PAGE_SIZE = 20;

const mergeRequests = (current: PublicRequest[], incoming: PublicRequest[]) => {
  const requests = new Map(current.map((request) => [request.id, request]));
  incoming.forEach((request) => {
    requests.set(request.id, request);
  });
  return Array.from(requests.values());
};

const fetchPortalPage = async (url: string) => {
  try {
    const response = await fetch(url);
    if (!response.ok) return null;
    const payload = (await response.json()) as ApiResponse<PublicPortal>;
    return payload.data;
  } catch {
    return null;
  }
};

const buildUrl = ({
  page,
  portal,
  filters,
}: {
  page: number;
  portal: PublicPortal;
  filters: PublicPortalFilters;
}) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
    sort: filters.sort,
  });
  if (filters.boardId) params.set("boardId", filters.boardId);
  if (filters.search.trim()) params.set("search", filters.search.trim());
  if (filters.status) params.set("status", filters.status);
  return `/api/public-portal/${portal.slug}?${params.toString()}`;
};

const FeedbackListSkeleton = () => (
  <Box>
    {Array.from({ length: 4 }).map((_, index) => (
      <Box className="border-border/70 border-b-[0.5px] py-5" key={index}>
        <Flex align="start" className="gap-4">
          <Box className="bg-surface-muted size-9 animate-pulse rounded-full" />
          <Box className="min-w-0 flex-1 space-y-3">
            <Box className="bg-surface-muted h-4 w-40 animate-pulse rounded-full" />
            <Box className="bg-surface-muted h-5 w-2/3 animate-pulse rounded-full" />
            <Box className="bg-surface-muted h-4 w-full max-w-xl animate-pulse rounded-full" />
          </Box>
        </Flex>
      </Box>
    ))}
  </Box>
);

const FeedbackSearch = ({
  initialValue,
  onSubmit,
}: {
  initialValue: string;
  onSubmit: (search: string) => void;
}) => {
  const [value, setValue] = useState(initialValue);

  return (
    <Box
      as="form"
      className="min-w-0 flex-1 md:max-w-80"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit(value.trim());
      }}
    >
      <Input
        className="h-10"
        leftIcon={<SearchIcon className="h-4" />}
        onChange={(event) => {
          setValue(event.target.value);
        }}
        placeholder="Search feedback..."
        type="search"
        value={value}
        variant="solid"
      />
    </Box>
  );
};

export const PublicFeedbackList = ({
  filters,
  onFiltersChange,
  portal,
}: {
  filters: PublicPortalFilters;
  onFiltersChange: (updates: Partial<PublicPortalFilters>) => void;
  portal: PublicPortal;
}) => {
  const [loadedRequests, setLoadedRequests] = useState<PublicRequest[] | null>(
    null,
  );
  const requests = loadedRequests ?? portal.requests;
  const [loadedHasMore, setLoadedHasMore] = useState<boolean | null>(null);
  const hasMore = loadedHasMore ?? portal.requestsHasMore;
  const [page, setPage] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const requestSequenceRef = useRef(0);
  const filterKey = [
    filters.boardId ?? "",
    filters.search,
    filters.sort,
    filters.status ?? "",
  ].join(":");
  const previousFilterKeyRef = useRef(filterKey);

  const loadPage = useEffectEvent(
    async (nextPage: number, mode: "append" | "replace") => {
      const requestSequence = requestSequenceRef.current + 1;
      requestSequenceRef.current = requestSequence;
      setIsLoading(true);
      const nextPortal = await fetchPortalPage(
        buildUrl({ filters, page: nextPage, portal }),
      );
      if (requestSequence !== requestSequenceRef.current) return;
      if (!nextPortal) {
        toast.error("Unable to load feedback");
        setIsLoading(false);
        return;
      }
      setLoadedRequests((current) =>
        mode === "append"
          ? mergeRequests(current ?? portal.requests, nextPortal.requests)
          : nextPortal.requests,
      );
      setLoadedHasMore(nextPortal.requestsHasMore);
      setPage(nextPage);
      setIsLoading(false);
    },
  );

  useEffect(() => {
    if (previousFilterKeyRef.current === filterKey) return;

    previousFilterKeyRef.current = filterKey;
    void loadPage(1, "replace");
  }, [filterKey]); // eslint-disable-line react-hooks/exhaustive-deps -- Effect Events must not be dependencies.

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasMore) return;

    const observer = new IntersectionObserver((entries) => {
      if (entries[0]?.isIntersecting && !isLoading) {
        void loadPage(page + 1, "append");
      }
    });
    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [hasMore, isLoading, page]); // eslint-disable-line react-hooks/exhaustive-deps -- Effect Events must not be dependencies.

  let feedbackContent: ReactNode = requests.map((request) => (
    <PublicRequestCard key={request.id} portal={portal} request={request} />
  ));
  if (isLoading && page === 1) {
    feedbackContent = <FeedbackListSkeleton />;
  } else if (requests.length === 0) {
    const selectedBoard = portal.boards.find(
      (board) => board.id === filters.boardId,
    );
    feedbackContent = (
      <Flex
        align="center"
        className="mt-28 min-h-72 text-center"
        direction="column"
        justify="center"
      >
        <Flex
          align="center"
          className="bg-surface-muted text-text-muted mb-4 size-12 rounded-xl"
          justify="center"
        >
          <RequestsIcon className="h-5 text-current" />
        </Flex>
        <Text className="text-[1.05rem]" fontWeight="semibold">
          No feedback yet
        </Text>
        <Text className="mt-1 max-w-sm" color="muted">
          {selectedBoard
            ? `No feedback has been submitted to ${selectedBoard.name}.`
            : "New feedback will appear here once it has been submitted."}
        </Text>
      </Flex>
    );
  }

  const changeFilters = (updates: Partial<PublicPortalFilters>) => {
    const changed = (
      Object.keys(updates) as (keyof PublicPortalFilters)[]
    ).some((key) => filters[key] !== updates[key]);

    if (changed) {
      setLoadedRequests([]);
      setLoadedHasMore(false);
      setPage(1);
      setIsLoading(true);
    }
    onFiltersChange(updates);
  };

  return (
    <Flex className="min-h-0 md:h-full" direction="column">
      <Box className="border-border/60 bg-background sticky top-0 z-10 shrink-0 border-b">
        <Flex align="center" className="gap-3 py-3" justify="between">
          <FeedbackSearch
            initialValue={filters.search}
            key={filters.search}
            onSubmit={(search) => {
              changeFilters({ search });
            }}
          />
          <Flex
            align="center"
            className="bg-surface-muted/85 h-10 shrink-0 gap-1 rounded-xl p-1"
          >
            {(["top", "newest", "oldest"] as const).map((option) => (
              <button
                className={cn(
                  "text-text-muted hover:text-foreground flex h-full items-center gap-1.5 rounded-xl border border-transparent px-3 capitalize transition",
                  {
                    "border-border bg-surface-elevated text-foreground":
                      filters.sort === option,
                  },
                )}
                key={option}
                onClick={() => {
                  changeFilters({ sort: option });
                }}
                type="button"
              >
                {option === "top" ? (
                  <ArrowUpDownIcon className="h-4 text-current" />
                ) : null}
                {option}
              </button>
            ))}
          </Flex>
        </Flex>
        <Flex
          align="center"
          className="bg-surface-muted/85 mb-3 h-10 w-full gap-1 overflow-x-auto rounded-xl p-1"
        >
          <button
            className={cn(
              "text-text-muted hover:text-foreground flex h-full min-w-max flex-1 items-center justify-center rounded-xl border border-transparent px-3.5 transition",
              {
                "border-border bg-surface-elevated text-foreground":
                  !filters.status,
              },
            )}
            onClick={() => {
              changeFilters({ status: undefined });
            }}
            type="button"
          >
            All
          </button>
          {requestFilters.map((filter) => {
            const meta = requestStatusMeta[filter];
            return (
              <button
                className={cn(
                  "text-text-muted hover:text-foreground flex h-full min-w-max flex-1 shrink-0 items-center justify-center gap-2 rounded-xl border border-transparent px-3.5 transition",
                  {
                    "border-border bg-surface-elevated text-foreground":
                      filters.status === filter,
                  },
                )}
                key={filter}
                onClick={() => {
                  changeFilters({ status: filter });
                }}
                type="button"
              >
                <span
                  className={cn("size-2 rounded-full", meta.dotClassName)}
                />
                <span>{meta.label}</span>
              </button>
            );
          })}
        </Flex>
      </Box>

      <Box className="min-h-0 md:flex-1 md:overflow-y-auto">
        {feedbackContent}
        <div ref={sentinelRef} />
        {isLoading && page > 1 ? <FeedbackListSkeleton /> : null}
      </Box>
    </Flex>
  );
};
