"use client";

import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { ArrowUpDownIcon, RequestsIcon, SearchIcon } from "icons";
import { Box, Flex, Input, Text } from "ui";
import { cn } from "lib";
import { toast } from "sonner";
import { requestFilters, requestStatusMeta } from "./status";
import { PublicRequestCard } from "./request-card";
import type { PublicPortal, PublicPortalFilters, PublicRequest } from "./types";
import { usePublicFeedbackList } from "./client-query";

const mergeRequests = (current: PublicRequest[], incoming: PublicRequest[]) => {
  const requests = new Map(current.map((request) => [request.id, request]));
  incoming.forEach((request) => {
    requests.set(request.id, request);
  });
  return Array.from(requests.values());
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
  initialFilters,
  onFiltersChange,
  portal,
}: {
  filters: PublicPortalFilters;
  initialFilters: PublicPortalFilters;
  onFiltersChange: (updates: Partial<PublicPortalFilters>) => void;
  portal: PublicPortal;
}) => {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isPending,
  } = usePublicFeedbackList({
    filters,
    initialFilters,
    initialPortal: portal,
  });
  const requests = (data?.pages ?? []).reduce<PublicRequest[]>(
    (current, page) => mergeRequests(current, page.requests),
    [],
  );
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (isError) toast.error("Unable to load feedback");
  }, [isError]);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasNextPage) return;

    const observer = new IntersectionObserver((entries) => {
      if (entries[0]?.isIntersecting && !isFetchingNextPage) {
        void fetchNextPage();
      }
    });
    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

  let feedbackContent: ReactNode = requests.map((request) => (
    <PublicRequestCard key={request.id} portal={portal} request={request} />
  ));
  if (isPending) {
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

      <Box className="hide-scrollbar min-h-0 md:flex-1 md:overflow-y-auto">
        {feedbackContent}
        <div ref={sentinelRef} />
        {isFetchingNextPage ? <FeedbackListSkeleton /> : null}
      </Box>
    </Flex>
  );
};
