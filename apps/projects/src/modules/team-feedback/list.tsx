"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useParams } from "next/navigation";
import { parseAsString, parseAsStringLiteral, useQueryState } from "nuqs";
import { useIntersectionObserver } from "react-intersection-observer-hook";
import { Box, Button, Flex, Text } from "ui";
import { TeamFeedbackCard } from "./card";
import { TeamFeedbackHeader } from "./header";
import { useTeamFeedbackInfinite } from "./hooks/use-team-feedback";
import type { TeamFeedbackListStatus } from "./types";

const emptyStateCopy: Record<TeamFeedbackListStatus, string> = {
  active:
    "Active customer ideas that have not been completed or closed will appear here.",
  pending: "There is no new feedback waiting for review.",
  reviewing: "No feedback is currently being reviewed.",
  planned: "No feedback has been planned yet.",
  in_progress: "No feedback is currently in progress.",
  completed: "No feedback has been completed yet.",
  closed: "No feedback has been closed.",
  trashed: "Deleted feedback will remain available here for 30 days.",
};

const feedbackListStatuses: TeamFeedbackListStatus[] = [
  "active",
  "pending",
  "reviewing",
  "planned",
  "in_progress",
  "completed",
  "closed",
  "trashed",
];

export const ListTeamFeedback = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const [status, setStatus] = useQueryState(
    "status",
    parseAsStringLiteral(feedbackListStatuses).withDefault("active"),
  );
  const [search, setSearch] = useQueryState(
    "search",
    parseAsString.withDefault(""),
  );
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isPending,
    refetch,
  } = useTeamFeedbackInfinite(teamId, status, search);
  const feedback = data?.pages.flatMap((page) => page.feedback) ?? [];
  const [triggerRef, { entry }] = useIntersectionObserver({
    threshold: 0,
    rootMargin: "240px",
  });

  useEffect(() => {
    if (entry?.isIntersecting && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [entry?.isIntersecting, fetchNextPage, hasNextPage, isFetchingNextPage]);

  let content: ReactNode;
  if (isPending) {
    content = (
      <Box className="space-y-0">
        {Array.from({ length: 4 }).map((_, index) => (
          <Box className="border-border border-b-[0.5px] px-5 py-4" key={index}>
            <Box className="bg-surface-muted mb-3 h-4 w-3/4 rounded" />
            <Box className="bg-surface-muted h-3 w-1/2 rounded" />
          </Box>
        ))}
      </Box>
    );
  } else if (isError) {
    content = (
      <Flex align="center" className="h-full px-6" justify="center">
        <Box>
          <Text align="center" className="mb-2" fontSize="xl">
            Couldn&apos;t load feedback
          </Text>
          <Text align="center" className="mb-4" color="muted">
            Check your connection and try again.
          </Text>
          <Flex justify="center">
            <Button
              color="tertiary"
              onClick={() => {
                void refetch();
              }}
              size="sm"
              variant="outline"
            >
              Try again
            </Button>
          </Flex>
        </Box>
      </Flex>
    );
  } else {
    content = (
      <>
        {feedback.map((item, index) => (
          <TeamFeedbackCard feedback={item} index={index} key={item.id} />
        ))}
        {hasNextPage ? <div className="h-4 w-full" ref={triggerRef} /> : null}
        {isFetchingNextPage ? (
          <Box className="space-y-0">
            {Array.from({ length: 2 }).map((_, index) => (
              <Box
                className="border-border border-b-[0.5px] px-5 py-4"
                key={index}
              >
                <Box className="bg-surface-muted mb-3 h-4 w-3/4 rounded" />
                <Box className="bg-surface-muted h-3 w-1/2 rounded" />
              </Box>
            ))}
          </Box>
        ) : null}
        {feedback.length === 0 ? (
          <Flex align="center" className="h-full px-6" justify="center">
            <Box>
              <Text align="center" className="mb-3" fontSize="xl">
                No feedback
              </Text>
              <Text align="center" color="muted">
                {search
                  ? `No feedback matches “${search}”.`
                  : emptyStateCopy[status]}
              </Text>
            </Box>
          </Flex>
        ) : null}
      </>
    );
  }

  return (
    <Box className="border-border/60 h-dvh border-r-[0.5px] pb-6">
      <TeamFeedbackHeader
        onSearchChange={(nextSearch) => {
          void setSearch(nextSearch || null);
        }}
        onStatusChange={(nextStatus) => {
          void setStatus(nextStatus);
        }}
        search={search}
        status={status}
      />
      <Box className="h-[calc(100dvh-4rem)] overflow-y-auto">{content}</Box>
    </Box>
  );
};
