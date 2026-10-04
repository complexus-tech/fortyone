"use client";

import { Box, Button, Text } from "ui";
import type { useViewsDirectory } from "./use-views-directory";

export const ViewsDirectoryStatus = ({
  directory,
}: {
  directory: ReturnType<typeof useViewsDirectory>;
}) => (
  <Box className="space-y-2">
    {directory.pending ? (
      <Text color="muted" role="status">
        Loading views...
      </Text>
    ) : null}
    {directory.teamsError ? (
      <Box className="space-y-2">
        <Text color="danger" role="alert">
          Available teams could not be loaded.
        </Text>
        <Button
          color="tertiary"
          onClick={directory.retryTeams}
          size="sm"
          variant="outline"
        >
          Try again
        </Button>
      </Box>
    ) : null}
    {directory.failedFeeds.map((feed) => (
      <Box className="space-y-2" key={feed.team.id}>
        <Text color="danger" role="alert">
          Views from {feed.team.name} could not be loaded.
        </Text>
        <Button
          color="tertiary"
          disabled={feed.isFetching}
          onClick={() => void feed.retry()}
          size="sm"
          variant="outline"
        >
          Try {feed.team.name} again
        </Button>
      </Box>
    ))}
    {directory.feeds
      .filter(
        (feed) => feed.isFetchedAfterMount && !feed.isError && feed.hasMore,
      )
      .map((feed) => (
        <Button
          color="tertiary"
          key={feed.team.id}
          loading={feed.isFetching}
          onClick={feed.loadMore}
          size="sm"
          variant="naked"
        >
          Load more from {feed.team.name}
        </Button>
      ))}
    {directory.hasMoreTeams ? (
      <Button
        color="tertiary"
        loading={directory.loadingMoreTeams}
        onClick={directory.loadMoreTeams}
        size="sm"
        variant="naked"
      >
        Load more teams
      </Button>
    ) : null}
  </Box>
);
