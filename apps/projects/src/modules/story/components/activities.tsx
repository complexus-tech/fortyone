import { Avatar, Box, Flex, Tabs, Text, Button, Skeleton } from "ui";
import { ClockIcon, CommentIcon } from "icons";
import { useSession } from "next-auth/react";
import { Activity } from "@/components/ui";
import { useStoryActivitiesInfinite } from "@/modules/story/hooks/story-activities";
import { CommentInput } from "./comment-input";
import { Comments } from "./comments";

export const Activities = ({
  className,
  storyId,
  teamId,
}: {
  className?: string;
  storyId: string;
  teamId: string;
}) => {
  const { data: session } = useSession();
  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useStoryActivitiesInfinite(storyId);

  const allActivities =
    infiniteData?.pages.flatMap((page) => page.activities) ?? [];

  const handleLoadMore = () => {
    fetchNextPage();
  };

  return (
    <Box className={className}>
      <Text
        as="h4"
        className="mb-4 flex items-center gap-1"
        fontWeight="medium"
      >
        <ClockIcon className="relative -top-px" />
        Activity feed
      </Text>

      <Tabs defaultValue="comments">
        <Tabs.List className="mx-0 mb-5 md:mx-0">
          <Tabs.Tab
            className="gap-1 px-2"
            leftIcon={<CommentIcon className="h-[1.1rem]" />}
            value="comments"
          >
            Comments
          </Tabs.Tab>
          <Tabs.Tab
            className="gap-1 px-2"
            leftIcon={<ClockIcon className="h-[1.1rem]" />}
            value="updates"
          >
            Updates
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="updates">
          <Flex direction="column">
            {allActivities.length === 0 ? (
              <Text>No updates available</Text>
            ) : (
              <>
                {allActivities.map((activity) => (
                  <Activity key={activity.id} {...activity} teamId={teamId} />
                ))}
                {isFetchingNextPage ? (
                  <Box className="mt-4 space-y-3">
                    {Array.from({ length: 2 }).map((_, i) => (
                      <Box className="flex gap-3" key={i}>
                        <Skeleton className="h-8 w-8 rounded-full" />
                        <Box className="flex-1 space-y-2">
                          <Skeleton className="h-4 w-32" />
                          <Skeleton className="h-4 w-48" />
                        </Box>
                      </Box>
                    ))}
                  </Box>
                ) : null}
              </>
            )}
          </Flex>
          {hasNextPage ? (
            <Box className="mt-2">
              <Button
                className="ml-6 px-3 text-[0.95rem]"
                color="tertiary"
                disabled={isFetchingNextPage}
                onClick={handleLoadMore}
                size="sm"
                variant="naked"
              >
                {isFetchingNextPage ? "Loading..." : "Load more updates"}
              </Button>
            </Box>
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel value="comments">
          <Flex align="start" className="mb-3">
            <Box className="z-1 flex aspect-square items-center rounded-full bg-white p-[0.3rem] bg-surface">
              <Avatar
                name={session?.user?.name ?? undefined}
                size="xs"
                src={session?.user?.image ?? undefined}
              />
            </Box>
            <CommentInput storyId={storyId} teamId={teamId} />
          </Flex>
          <Comments storyId={storyId} teamId={teamId} />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
