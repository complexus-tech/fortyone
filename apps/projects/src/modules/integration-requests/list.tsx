"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { useIntersectionObserver } from "react-intersection-observer-hook";
import { Box, Flex, Text } from "ui";
import { IntegrationRequestCard } from "./card";
import { IntegrationRequestsHeader } from "./header";
import { useTeamIntegrationRequestsInfinite } from "./hooks/use-team-requests";

export const ListIntegrationRequests = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending } =
    useTeamIntegrationRequestsInfinite(teamId);
  const requests = data?.pages.flatMap((page) => page.requests) ?? [];
  const [triggerRef, { entry }] = useIntersectionObserver({
    threshold: 0,
    rootMargin: "240px",
  });

  useEffect(() => {
    if (entry?.isIntersecting && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [entry?.isIntersecting, fetchNextPage, hasNextPage, isFetchingNextPage]);

  return (
    <Box className="border-border/60 h-dvh border-r-[0.5px] pb-6">
      <IntegrationRequestsHeader
        requestCount={requests.length}
        teamId={teamId}
      />
      <Box className="h-[calc(100dvh-4rem)] overflow-y-auto">
        {isPending ? (
          <Box className="space-y-0">
            {Array.from({ length: 4 }).map((_, index) => (
              <Box
                className="border-border border-b-[0.5px] px-5 py-4"
                key={index}
              >
                <Box className="bg-surface-muted mb-3 h-4 w-3/4 rounded" />
                <Box className="bg-surface-muted h-3 w-1/2 rounded" />
              </Box>
            ))}
          </Box>
        ) : (
          <>
            {requests.map((request, index) => (
              <IntegrationRequestCard
                index={index}
                key={request.id}
                request={request}
              />
            ))}
            {hasNextPage ? (
              <div className="h-4 w-full" ref={triggerRef} />
            ) : null}
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
            {requests.length === 0 ? (
              <Flex align="center" className="h-full px-6" justify="center">
                <Box>
                  <Text align="center" className="mb-3" fontSize="xl">
                    No requests
                  </Text>
                  <Text align="center" color="muted">
                    New integration items will appear here before becoming
                    stories.
                  </Text>
                </Box>
              </Flex>
            ) : null}
          </>
        )}
      </Box>
    </Box>
  );
};
