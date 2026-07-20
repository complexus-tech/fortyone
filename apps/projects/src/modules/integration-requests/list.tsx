"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { parseAsString, useQueryState } from "nuqs";
import { useIntersectionObserver } from "react-intersection-observer-hook";
import { Box, Flex, Text } from "ui";
import { IntegrationRequestCard } from "./card";
import { IntegrationRequestsHeader } from "./header";
import { useTeamIntegrationRequestsInfinite } from "./hooks/use-team-requests";

export const ListIntegrationRequests = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const [search, setSearch] = useQueryState(
    "search",
    parseAsString.withDefault(""),
  );
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending } =
    useTeamIntegrationRequestsInfinite(teamId, "pending", search);
  const requests = data?.pages.flatMap((page) => page.requests) ?? [];
  const totalCount = data?.pages[0]?.pagination.totalCount ?? 0;
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
        onSearchChange={(nextSearch) => {
          void setSearch(nextSearch || null);
        }}
        requestCount={totalCount}
        search={search}
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
                    {search ? "No matching intake" : "No intake items"}
                  </Text>
                  <Text align="center" color="muted">
                    {search
                      ? `No intake items match “${search}”.`
                      : "New integration items will appear here before becoming stories."}
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
