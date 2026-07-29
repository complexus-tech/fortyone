"use client";

import { BellIcon, NotificationsCheckIcon } from "icons";
import { Button, Tooltip } from "ui";
import { useStoryById } from "@/modules/story/hooks/story";
import { useSetStoryWatchingMutation } from "../hooks/collaboration-mutations";

export const StorySubscriptionButton = ({ storyId }: { storyId: string }) => {
  const { data } = useStoryById(storyId);
  const subscriptionMutation = useSetStoryWatchingMutation();

  if (!data) {
    return null;
  }

  const { deletedAt, isWatching: isSubscribed } = data;
  const label = isSubscribed ? "Unsubscribe" : "Subscribe";

  return (
    <Tooltip title={label}>
      <Button
        active={isSubscribed}
        aria-label={label}
        asIcon
        color="tertiary"
        disabled={Boolean(deletedAt) || subscriptionMutation.isPending}
        onClick={() => {
          subscriptionMutation.mutate({
            storyId,
            watching: !isSubscribed,
          });
        }}
        type="button"
        variant="naked"
      >
        {isSubscribed ? (
          <NotificationsCheckIcon className="h-5 w-auto" />
        ) : (
          <BellIcon className="h-5 w-auto" />
        )}
      </Button>
    </Tooltip>
  );
};
