"use client";

import type { ReactNode } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import { ArrowUpDownIcon, RequestsIcon, SearchIcon } from "icons";
import { Box, Flex, Input, Text } from "ui";
import { cn } from "lib";
import { toast } from "sonner";
import { requestFilters, requestStatusMeta } from "./status";
import { PublicRequestCard } from "./request-card";
import type { PublicPortal, PublicRequest, PublicRequestStatus } from "./types";

type SortMode = "top" | "newest" | "oldest";

type ApiResponse<T> = {
  data: T;
};

const PAGE_SIZE = 20;

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
  boardId,
  search,
  sort,
  status,
}: {
  page: number;
  portal: PublicPortal;
  boardId?: string;
  search: string;
  sort: SortMode;
  status?: PublicRequestStatus;
}) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
    sort,
  });
  if (boardId) params.set("boardId", boardId);
  if (search.trim()) params.set("search", search.trim());
  if (status) params.set("status", status);
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

export const PublicFeedbackList = ({
  boardId,
  portal,
}: {
  boardId?: string;
  portal: PublicPortal;
}) => {
  const [loadedRequests, setLoadedRequests] = useState<PublicRequest[] | null>(
    null,
  );
  const requests = loadedRequests ?? portal.requests;
  const [hasMore, setHasMore] = useState(portal.requestsHasMore);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<PublicRequestStatus | undefined>();
  const [sort, setSort] = useState<SortMode>("top");
  const [search, setSearch] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const hasMountedRef = useRef(false);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const loadPage = useCallback(
    async (nextPage: number, mode: "append" | "replace") => {
      setIsLoading(true);
      const nextPortal = await fetchPortalPage(
        buildUrl({ boardId, page: nextPage, portal, search, sort, status }),
      );
      if (!nextPortal) {
        toast.error("Unable to load feedback");
        setIsLoading(false);
        return;
      }
      setLoadedRequests((current) =>
        mode === "append"
          ? [...(current ?? portal.requests), ...nextPortal.requests]
          : nextPortal.requests,
      );
      setHasMore(nextPortal.requestsHasMore);
      setPage(nextPage);
      setIsLoading(false);
    },
    [boardId, portal, search, sort, status],
  );

  useEffect(() => {
    if (!hasMountedRef.current) {
      hasMountedRef.current = true;
      return;
    }

    void loadPage(1, "replace");
  }, [loadPage]);

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
  }, [hasMore, isLoading, loadPage, page]);

  let feedbackContent: ReactNode = requests.map((request) => (
    <PublicRequestCard key={request.id} portal={portal} request={request} />
  ));
  if (isLoading && page === 1) {
    feedbackContent = <FeedbackListSkeleton />;
  } else if (requests.length === 0) {
    const selectedBoard = portal.boards.find((board) => board.id === boardId);
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

  return (
    <>
      <Box className="border-border/60 border-b">
        <Flex align="center" className="gap-3 py-3" justify="between">
          <Box className="min-w-0 flex-1 md:max-w-72">
            <Input
              className="h-10"
              leftIcon={<SearchIcon className="h-4" />}
              onChange={(event) => {
                setSearch(event.target.value);
              }}
              placeholder="Search feedback..."
              type="search"
              value={search}
              variant="solid"
            />
          </Box>
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
                      sort === option,
                  },
                )}
                key={option}
                onClick={() => {
                  setSort(option);
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
                "border-border bg-surface-elevated text-foreground": !status,
              },
            )}
            onClick={() => {
              setStatus(undefined);
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
                      status === filter,
                  },
                )}
                key={filter}
                onClick={() => {
                  setStatus(filter);
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

      <Box>
        {feedbackContent}
        <div ref={sentinelRef} />
        {isLoading && page > 1 ? <FeedbackListSkeleton /> : null}
      </Box>
    </>
  );
};
