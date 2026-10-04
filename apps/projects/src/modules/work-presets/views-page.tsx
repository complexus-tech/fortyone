"use client";

import { useEffect, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Box, Flex, Skeleton, Text } from "ui";
import { ViewsIcon } from "icons";
import { HeaderContainer } from "@/components/shared/header-container";
import { MobileMenuButton } from "@/components/shared/mobile-menu";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useViewsDirectory } from "./use-views-directory";
import { savedViewPath } from "./view-link";
import { ViewsDirectoryStatus } from "./views-directory-status";

const ViewsEntry = ({ teamId }: { teamId: string | null }) => {
  const directory = useViewsDirectory(teamId);
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const navigating = useRef(false);
  const first = directory.views.at(0)?.view ?? null;
  const initialLoading =
    directory.pending &&
    !first &&
    !directory.teamsError &&
    !directory.failedFeeds.length;
  const showSkeleton = initialLoading || Boolean(first);
  useEffect(() => {
    if (!first || navigating.current) return;
    navigating.current = true;
    router.replace(withWorkspace(savedViewPath(first.teamId, first.id)));
  }, [first, router, withWorkspace]);
  useEffect(() => {
    if (
      first ||
      directory.pending ||
      directory.teamsError ||
      directory.failedFeeds.length
    )
      return;
    const moreViews = directory.feeds.find(
      (feed) => feed.isFetchedAfterMount && feed.hasMore && !feed.isFetching,
    );
    if (moreViews) moreViews.loadMore();
    else if (directory.hasMoreTeams && !directory.loadingMoreTeams)
      directory.loadMoreTeams();
  }, [first, directory]);
  return (
    <Box className="flex h-full min-h-0 flex-col">
      <HeaderContainer className="justify-between">
        <Flex align="center" className="min-w-0" gap={2}>
          <MobileMenuButton />
          <ViewsIcon />
          <Text as="h1">Views</Text>
        </Flex>
        {showSkeleton ? (
          <Flex align="center" aria-hidden gap={2}>
            <Skeleton className="h-8 w-20" />
            <Skeleton className="h-8 w-16" />
            <Skeleton className="h-8 w-24" />
          </Flex>
        ) : null}
      </HeaderContainer>
      {showSkeleton ? (
        <>
          <Text className="sr-only" role="status">
            {first ? `Opening ${first.name}...` : "Loading views..."}
          </Text>
          <Box aria-hidden className="min-h-0 flex-1 overflow-hidden" inert>
            <BoardSkeleton className="h-full" layout="list" />
          </Box>
        </>
      ) : (
        <Box className="mx-auto flex w-full max-w-xl flex-1 flex-col items-center justify-center gap-4 p-6 text-center">
          {directory.complete && !first ? (
            <>
              <ViewsIcon className="text-text-muted h-8 w-auto" />
              <Text as="h2" fontWeight="medium">
                No saved views yet
              </Text>
              <Text color="muted">
                Apply filters in My Work or a team’s tasks, then choose Save as
                to keep that view here.
              </Text>
            </>
          ) : null}
          <ViewsDirectoryStatus directory={directory} />
        </Box>
      )}
    </Box>
  );
};

export const ViewsPage = () => {
  const teamId = useSearchParams().get("team");
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return (
    <ViewsEntry
      key={`${workspaceSlug}:${session?.user.id ?? ""}:${teamId ?? ""}`}
      teamId={teamId}
    />
  );
};
