"use client";

import type { ReactNode } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { ArrowUpIcon, StoryIcon } from "icons";
import { Avatar, Box, Flex, Text } from "ui";
import { cn } from "lib";
import { roadmapStatuses, requestStatusMeta } from "./status";
import type { PublicPortal, PublicRequest } from "./types";
import { getBoard, getRequestPath } from "./utils";
import { getPublicAvatarColor } from "./avatar-color";

type ApiResponse<T> = {
  data: T;
};

const PAGE_SIZE = 20;

const roadmapLabels = {
  planned: {
    title: "Planned",
    description: "Committed and queued",
  },
  in_progress: {
    title: "In Progress",
    description: "Actively being delivered",
  },
  completed: {
    title: "Done",
    description: "Recently completed",
  },
};

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
      "bg-surface-muted text-text-primary h-8 shrink-0 gap-0 rounded-lg px-2.5 text-[0.9rem] font-medium",
      className,
    )}
  >
    <ArrowUpIcon className="-mr-1.5 h-3.5 text-current" />
    <span>{voteCount}</span>
  </Flex>
);

const BoardLabel = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = getBoard(portal, request.boardId);
  const meta = requestStatusMeta[request.status];

  if (!board) return null;

  return (
    <Flex align="center" className="min-w-0" gap={2}>
      <span className={cn("size-2 shrink-0 rounded-full", meta.dotClassName)} />
      <Text
        className="truncate text-[0.94rem]"
        color="muted"
        fontWeight="semibold"
      >
        {board.name}
      </Text>
    </Flex>
  );
};

const RoadmapRequestCard = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => (
  <Link className="group block" href={getRequestPath(portal, request)}>
    <Box className="border-border bg-surface shadow-shadow hover:bg-surface-elevated w-full rounded-2xl border-[0.5px] p-4 shadow-sm backdrop-blur transition duration-200 ease-linear select-none">
      <BoardLabel portal={portal} request={request} />
      <Text
        className="mt-2 line-clamp-2 text-[1.05rem] leading-[1.4rem] group-hover:opacity-90"
        fontWeight="semibold"
      >
        {request.title}
      </Text>
      {request.roadmapSummary || request.description ? (
        <Text className="text-text-muted/80 mt-1.5 line-clamp-2 leading-6">
          {request.roadmapSummary || request.description}
        </Text>
      ) : null}
      <Flex align="center" className="mt-5 min-w-0 gap-3" justify="between">
        <Flex align="center" className="min-w-0 flex-1" gap={2}>
          <Avatar
            className="shrink-0"
            name={request.authorName}
            rounded="full"
            size="sm"
            src={request.authorAvatar}
            style={{
              backgroundColor: getPublicAvatarColor(request.authorName),
            }}
          />
          <Text className="truncate text-[0.95rem]" fontWeight="medium">
            {request.authorName}
          </Text>
        </Flex>
        <RoadmapVote voteCount={request.voteCount} />
      </Flex>
    </Box>
  </Link>
);

const ColumnSkeleton = () => (
  <>
    {Array.from({ length: 3 }).map((_, index) => (
      <Box
        className="border-border bg-surface w-full animate-pulse rounded-2xl border-[0.5px] p-4"
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
  portal,
  status,
}: {
  portal: PublicPortal;
  status: (typeof roadmapStatuses)[number];
}) => {
  const [items, setItems] = useState<PublicRequest[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [page, setPage] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const column = roadmapLabels[status];
  const meta = requestStatusMeta[status];

  const loadPage = useCallback(
    async (nextPage: number, mode: "append" | "replace") => {
      setIsLoading(true);
      try {
        const response = await fetch(
          buildStatusUrl({ page: nextPage, portal, status }),
        );
        if (!response.ok) return;
        const payload = (await response.json()) as ApiResponse<PublicPortal>;
        setItems((current) =>
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
    [portal, status],
  );

  useEffect(() => {
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

  let columnContent: ReactNode = items.map((request) => (
    <RoadmapRequestCard key={request.id} portal={portal} request={request} />
  ));
  if (isLoading && page === 1) {
    columnContent = <ColumnSkeleton />;
  } else if (items.length === 0) {
    columnContent = (
      <Flex align="center" className="h-56" justify="center">
        <Text color="muted">Nothing here yet</Text>
      </Flex>
    );
  }

  return (
    <Box className="flex min-h-0 min-w-0 flex-col" key={status}>
      <Flex align="center" className="h-14 shrink-0 px-1" justify="between">
        <Flex align="center" className="min-w-0" gap={2}>
          <span
            className={cn("size-2.5 shrink-0 rounded-full", meta.dotClassName)}
          />
          <Text className="truncate" fontWeight="medium">
            {column.title}
          </Text>
        </Flex>
        <Flex align="center" className="text-text-muted shrink-0 gap-1.5">
          <StoryIcon className="h-5 w-auto" strokeWidth={2} />
          <span>{items.length}</span>
        </Flex>
      </Flex>
      <Flex className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto rounded-2xl p-2 transition">
        {columnContent}
        <div ref={sentinelRef} />
        {isLoading && page > 1 ? <ColumnSkeleton /> : null}
      </Flex>
    </Box>
  );
};

export const RoadmapBoard = ({ portal }: { portal: PublicPortal }) => (
  <Box className="mx-auto w-full max-w-[78rem] px-4 pt-8 pb-6 md:px-6">
    <Box className="grid min-h-[calc(100dvh-11rem)] gap-5 md:min-h-[calc(100dvh-8.5rem)] md:grid-cols-3">
      {roadmapStatuses.map((status) => (
        <RoadmapColumn key={status} portal={portal} status={status} />
      ))}
    </Box>
  </Box>
);
