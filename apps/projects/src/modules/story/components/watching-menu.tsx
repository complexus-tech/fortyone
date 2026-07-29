"use client";

import { BellIcon } from "icons";
import { Avatar, Box, Button, Divider, Flex, Popover, Text } from "ui";
import { useStoryById } from "@/modules/story/hooks/story";
import { useSetStoryWatchingMutation } from "../hooks/collaboration-mutations";

const getWatchingDescription = (
  watchingReason: "assignee" | "collaborator" | "watcher" | null,
  isWatching: boolean,
) => {
  switch (watchingReason) {
    case "assignee":
      return isWatching
        ? "You receive updates because this story is assigned to you."
        : "Updates are muted. You remain the assignee.";
    case "collaborator":
      return isWatching
        ? "You receive updates because you collaborate on this story."
        : "Updates are muted. You remain a collaborator.";
    case "watcher":
      return "You chose to receive updates about this story.";
    default:
      return "Watch this story to receive meaningful updates without becoming a collaborator.";
  }
};

export const StoryWatchingMenu = ({ storyId }: { storyId: string }) => {
  const { data } = useStoryById(storyId);
  const watchingMutation = useSetStoryWatchingMutation();

  if (!data) {
    return null;
  }

  const { deletedAt, isWatching, watcherCount, watchers, watchingReason } =
    data;

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          active={isWatching}
          color="tertiary"
          leftIcon={<BellIcon className="h-5 w-auto" />}
          size="sm"
          type="button"
          variant="naked"
        >
          <span className="hidden xl:inline">
            {isWatching ? "Watching" : "Watch"}
          </span>
          {watcherCount > 0 ? (
            <span className="text-text-muted text-xs">{watcherCount}</span>
          ) : null}
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className="w-80">
        <Box className="px-3 py-1">
          <Text fontWeight="semibold">
            {isWatching ? "Watching this story" : "Watch this story"}
          </Text>
          <Text className="mt-1 text-sm" color="muted">
            {getWatchingDescription(watchingReason, isWatching)}
          </Text>
        </Box>
        {watchers.length > 0 ? (
          <>
            <Divider className="my-2" />
            <Box className="max-h-48 overflow-y-auto px-3 py-1">
              <Text
                className="mb-2 text-xs"
                color="muted"
                fontWeight="semibold"
                transform="uppercase"
              >
                Receiving updates
              </Text>
              <Flex direction="column" gap={2}>
                {watchers.map((watcher) => (
                  <Flex align="center" gap={2} key={watcher.id}>
                    <Avatar
                      name={watcher.fullName || watcher.username}
                      size="xs"
                      src={watcher.avatarUrl}
                    />
                    <Text className="truncate text-sm">
                      {watcher.fullName || watcher.username}
                    </Text>
                  </Flex>
                ))}
              </Flex>
            </Box>
          </>
        ) : null}
        <Divider className="my-2" />
        <Box className="px-3 py-1">
          <Button
            color={isWatching ? "tertiary" : "primary"}
            disabled={Boolean(deletedAt)}
            fullWidth
            loading={watchingMutation.isPending}
            loadingText={isWatching ? "Stopping..." : "Watching..."}
            onClick={() => {
              watchingMutation.mutate({
                storyId,
                watching: !isWatching,
              });
            }}
            size="sm"
            type="button"
            variant={isWatching ? "outline" : "solid"}
          >
            {isWatching ? "Stop watching" : "Watch story"}
          </Button>
        </Box>
      </Popover.Content>
    </Popover>
  );
};
