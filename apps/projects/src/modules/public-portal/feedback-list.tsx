"use client";

import type { ReactNode } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import { ArrowUpDownIcon, SearchIcon } from "icons";
import { Box, Flex, Input, Text } from "ui";
import { cn } from "lib";
import { requestFilters, requestStatusMeta } from "./status";
import { PublicRequestCard } from "./request-card";
import type { PublicPortal, PublicRequest, PublicRequestStatus } from "./types";

type SortMode = "top" | "newest" | "oldest";

type ApiResponse<T> = {
  data: T;
};

const PAGE_SIZE = 20;

const buildUrl = ({
  page,
  portal,
  search,
  sort,
  status,
}: {
  page: number;
  portal: PublicPortal;
  search: string;
  sort: SortMode;
  status?: PublicRequestStatus;
}) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
    sort,
  });
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

export const PublicFeedbackList = ({ portal }: { portal: PublicPortal }) => {
  const [requests, setRequests] = useState<PublicRequest[]>(portal.requests);
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
      try {
        const response = await fetch(
          buildUrl({ page: nextPage, portal, search, sort, status }),
        );
        if (!response.ok) return;
        const payload = (await response.json()) as ApiResponse<PublicPortal>;
        setRequests((current) =>
          mode === "append"
            ? [...current, ...payload.data.requests]
            : payload.data.requests,
        );
        setHasMore(payload.data.requestsHasMore);
        setPage(nextPage);
      } finally {
        setIsLoading(false);
      }
    },
    [portal, search, sort, status],
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
    feedbackContent = (
      <Flex align="center" className="min-h-56" justify="center">
        <Text color="muted">No feedback found</Text>
      </Flex>
    );
  }

  return (
    <>
      <Box className="border-border/60 bg-surface/20 border-b">
        <Box className="mx-auto flex min-h-16 max-w-[78rem] flex-wrap items-center gap-4 px-4 py-3 md:px-6">
          <Box className="w-full md:w-72">
            <Input
              className="h-10 rounded-full"
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
            className="bg-surface border-border/70 shadow-shadow/30 shrink-0 gap-1 rounded-full border p-1 shadow-sm"
          >
            <button
              className={cn(
                "text-text-muted hover:bg-state-hover hover:text-foreground rounded-full px-3.5 py-1.5 transition",
                {
                  "bg-state-selected/50 text-foreground dark:bg-state-selected shadow-xs":
                    !status,
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
                    "text-text-muted hover:bg-state-hover hover:text-foreground flex shrink-0 items-center gap-2 rounded-full px-3.5 py-1.5 transition",
                    {
                      "bg-state-selected/50 text-foreground dark:bg-state-selected shadow-xs":
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
          <Flex
            align="center"
            className="bg-surface border-border/70 shadow-shadow/30 shrink-0 gap-1 rounded-full border p-1 shadow-sm"
          >
            {(["top", "newest", "oldest"] as const).map((option) => (
              <button
                className={cn(
                  "text-text-muted hover:bg-state-hover hover:text-foreground flex items-center gap-1.5 rounded-full px-3.5 py-1.5 capitalize transition",
                  {
                    "bg-state-selected/50 text-foreground dark:bg-state-selected shadow-xs":
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
        </Box>
      </Box>

      <Box>
        {feedbackContent}
        <div ref={sentinelRef} />
        {isLoading && page > 1 ? <FeedbackListSkeleton /> : null}
      </Box>
    </>
  );
};
