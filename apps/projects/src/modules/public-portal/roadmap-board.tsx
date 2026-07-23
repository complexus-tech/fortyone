"use client";

import type { ReactNode } from "react";
import { useEffect, useRef } from "react";
import Link from "next/link";
import { useInfiniteQuery } from "@tanstack/react-query";
import { RoadmapIcon, ThumbsUpIcon } from "icons";
import { Avatar, Box, Flex, Text } from "ui";
import { cn } from "lib";
import { Dot } from "@/components/ui/dot";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { publicPortalKeys } from "./query-keys";
import { roadmapStatuses } from "./status";
import type { PublicPortal, PublicRequest } from "./types";
import { getAuthorPath, getBoard, getRequestPath } from "./utils";
import { getPublicAvatarColor } from "./avatar-color";
import { PublicBoardPill } from "./board-pill";

type ApiResponse<T> = {
  data: T;
};

const PAGE_SIZE = 20;

const roadmapLabels = {
  planned: {
    title: "Planned",
    description: "Committed and queued",
    emptyTitle: "Nothing planned yet",
    markerColor: "var(--color-primary)",
  },
  in_progress: {
    title: "In Progress",
    description: "Actively being delivered",
    emptyTitle: "Nothing in progress",
    markerColor: "var(--color-info)",
  },
  completed: {
    title: "Done",
    description: "Recently completed",
    emptyTitle: "Nothing completed yet",
    markerColor: "var(--color-success)",
  },
};

type RoadmapStatus = (typeof roadmapStatuses)[number];

const buildStatusUrl = ({
  page,
  portal,
  status,
}: {
  page: number;
  portal: PublicPortal;
  status: (typeof roadmapStatuses)[number];
}) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
    sort: "newest",
    status,
  });
  return `/api/public-portal/${portal.slug}?${params.toString()}`;
};

const fetchRoadmapPage = async ({
  page,
  portal,
  status,
}: {
  page: number;
  portal: PublicPortal;
  status: RoadmapStatus;
}) => {
  const response = await fetch(buildStatusUrl({ page, portal, status }));

  if (!response.ok) {
    throw new Error(`Unable to load the ${status} roadmap column`);
  }

  const payload = (await response.json()) as ApiResponse<PublicPortal>;
  return payload.data;
};

const useRoadmapColumn = (portal: PublicPortal, status: RoadmapStatus) => {
  const query = useInfiniteQuery({
    queryKey: publicPortalKeys.roadmap(portal.slug, status),
    queryFn: ({ pageParam }) =>
      fetchRoadmapPage({ page: pageParam, portal, status }),
    getNextPageParam: (lastPage, _pages, lastPageParam) =>
      lastPage.requestsHasMore ? lastPageParam + 1 : undefined,
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });

  return {
    hasMore: query.hasNextPage,
    isError: query.isError,
    isLoadingMore: query.isFetchingNextPage,
    isPending: query.isPending,
    items: query.data?.pages.flatMap((page) => page.requests) ?? [],
    loadMore: query.fetchNextPage,
  };
};

type RoadmapColumnState = ReturnType<typeof useRoadmapColumn>;

const RoadmapVote = ({
  className,
  voteCount,
}: {
  className?: string;
  voteCount: number;
}) => (
  <Flex
    align="center"
    className={cn(
      "text-text-muted h-7 shrink-0 gap-1 text-[0.9rem] font-medium",
      className,
    )}
  >
    <ThumbsUpIcon className="h-3.5 text-current" strokeWidth={2} />
    <span>{voteCount}</span>
  </Flex>
);

const BoardBadge = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = getBoard(portal, request.boardId);

  if (!board) return null;

  return <PublicBoardPill board={board} />;
};

const RoadmapAuthorAvatar = ({ request }: { request: PublicRequest }) => (
  <span className="border-border bg-surface-elevated flex size-6.5 shrink-0 items-center justify-center rounded-lg border-[0.5px] p-0.5">
    <Avatar
      name={request.authorName}
      rounded="md"
      size="xs"
      src={request.authorAvatar}
      style={{
        backgroundColor: getPublicAvatarColor(request.authorName),
      }}
    />
  </span>
);

const RoadmapRequestCard = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const authorPath = getAuthorPath(portal, request.authorId);

  return (
    <Box className="border-border/70 bg-surface/50 shadow-shadow hover:bg-surface w-full rounded-xl border-[0.5px] p-4 shadow-sm backdrop-blur transition duration-200 ease-linear select-none">
      <Link className="group block" href={getRequestPath(portal, request)}>
        <Text
          className="line-clamp-3 text-[1.1rem] leading-[1.4rem] group-hover:opacity-90"
          fontWeight="medium"
        >
          {request.title}
        </Text>
        {request.roadmapSummary || request.description ? (
          <Text className="text-text-muted/80 mt-1.5 line-clamp-2 leading-6">
            {request.roadmapSummary || request.description}
          </Text>
        ) : null}
      </Link>
      <Flex align="end" className="mt-3 min-w-0 gap-3" justify="between">
        <Flex align="center" className="min-w-0 flex-1 gap-1.5" wrap>
          {authorPath ? (
            <Link
              aria-label={`View ${request.authorName}'s profile`}
              className="focus-visible:ring-ring shrink-0 rounded-lg transition-opacity hover:opacity-80 focus-visible:ring-2 focus-visible:outline-none"
              href={authorPath}
            >
              <RoadmapAuthorAvatar request={request} />
            </Link>
          ) : (
            <RoadmapAuthorAvatar request={request} />
          )}
          <BoardBadge portal={portal} request={request} />
        </Flex>
        <RoadmapVote voteCount={request.voteCount} />
      </Flex>
    </Box>
  );
};

const ColumnSkeleton = () => (
  <>
    {Array.from({ length: 3 }).map((_, index) => (
      <Box
        className="border-border bg-surface w-full animate-pulse rounded-xl border-[0.5px] p-4"
        key={index}
      >
        <Box className="bg-surface-muted h-4 w-28 rounded-full" />
        <Box className="bg-surface-muted mt-3 h-5 w-3/4 rounded-full" />
        <Box className="bg-surface-muted mt-2 h-4 w-full rounded-full" />
      </Box>
    ))}
  </>
);

const RoadmapColumn = ({
  columnState,
  portal,
  status,
}: {
  columnState: RoadmapColumnState;
  portal: PublicPortal;
  status: RoadmapStatus;
}) => {
  const { hasMore, isError, isLoadingMore, isPending, items, loadMore } =
    columnState;
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const column = roadmapLabels[status];
  const isEmpty = !isPending && !isError && items.length === 0;

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasMore) return;
    const observer = new IntersectionObserver((entries) => {
      if (entries[0]?.isIntersecting && !isLoadingMore) {
        loadMore();
      }
    });
    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [hasMore, isLoadingMore, loadMore]);

  let columnContent: ReactNode = items.map((request) => (
    <RoadmapRequestCard key={request.id} portal={portal} request={request} />
  ));
  if (isPending) {
    columnContent = <ColumnSkeleton />;
  } else if (isError) {
    columnContent = (
      <Flex align="center" className="h-56 px-4 text-center" justify="center">
        <Text color="muted">Unable to load roadmap items.</Text>
      </Flex>
    );
  } else if (items.length === 0) {
    columnContent = (
      <Flex
        align="center"
        className="min-h-56 flex-1 px-4 text-center"
        justify="center"
      >
        <Text color="muted">{column.emptyTitle}</Text>
      </Flex>
    );
  }

  return (
    <Box className="flex min-h-0 min-w-0 flex-col" key={status}>
      <Box className="shrink-0 px-1 pt-2 pb-3">
        <Flex align="center" justify="between">
          <Flex align="center" className="min-w-0" gap={2}>
            <Dot className="size-2.5" color={column.markerColor} />
            <Text as="h2" className="truncate" fontWeight="medium">
              {column.title}
            </Text>
          </Flex>
          <Text className="shrink-0 tabular-nums" color="muted">
            {items.length}
          </Text>
        </Flex>
        <Text className="mt-1.5 text-base leading-5" color="muted">
          {column.description}
        </Text>
      </Box>
      <Flex
        className={cn(
          "flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto rounded-xl p-2",
          {
            "border-border/60 bg-surface/20 min-h-56 border border-dashed":
              isEmpty,
          },
        )}
      >
        {columnContent}
        <div ref={sentinelRef} />
        {isLoadingMore ? <ColumnSkeleton /> : null}
      </Flex>
    </Box>
  );
};

const RoadmapEmptyState = () => (
  <Flex
    align="center"
    className="min-h-56 flex-1 px-4 text-center"
    direction="column"
    justify="center"
  >
    <RoadmapIcon className="text-text-muted h-16 w-auto" strokeWidth={1.3} />
    <Text as="h2" className="mt-7" fontSize="3xl">
      Nothing is on the roadmap yet
    </Text>
    <Text className="mt-3 max-w-md" color="muted">
      Feedback will appear here once the team plans it for delivery.
    </Text>
  </Flex>
);

export const RoadmapBoard = ({ portal }: { portal: PublicPortal }) => {
  const planned = useRoadmapColumn(portal, "planned");
  const inProgress = useRoadmapColumn(portal, "in_progress");
  const completed = useRoadmapColumn(portal, "completed");
  const columns: Record<RoadmapStatus, ReturnType<typeof useRoadmapColumn>> = {
    completed,
    in_progress: inProgress,
    planned,
  };
  const hasLoadedEveryColumn = roadmapStatuses.every(
    (status) => !columns[status].isPending,
  );
  const isRoadmapEmpty =
    hasLoadedEveryColumn &&
    roadmapStatuses.every(
      (status) =>
        !columns[status].isError && columns[status].items.length === 0,
    );

  return (
    <Box className="mx-auto flex min-h-full w-full max-w-[78rem] flex-col px-4 pt-8 md:h-full md:min-h-0 md:overflow-hidden md:px-6">
      {isRoadmapEmpty ? (
        <RoadmapEmptyState />
      ) : (
        <Box className="grid min-h-0 flex-1 gap-5 md:grid-cols-3">
          {roadmapStatuses.map((status) => (
            <RoadmapColumn
              columnState={columns[status]}
              key={status}
              portal={portal}
              status={status}
            />
          ))}
        </Box>
      )}
    </Box>
  );
};
