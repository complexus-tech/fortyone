import { Avatar, Box, Flex, Tabs, Text } from "ui";
import { ClockIcon, CommentIcon } from "icons";
import { useSession } from "next-auth/react";
import { Activity } from "@/components/ui";
import type { StoryActivity } from "@/modules/stories/types";
import { CommentInput } from "./comment-input";
import { Comments } from "./comments";

export const Activities = ({
  activities,
  className,
  storyId,
}: {
  activities: StoryActivity[];
  className?: string;
  storyId: string;
}) => {
  const { data: session } = useSession();

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
        <Tabs.List className="mx-0 mb-5">
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
            {activities.length === 0 ? (
              <Text>No updates available</Text>
            ) : (
              activities.map((activity) => (
                <Activity key={activity.id} {...activity} />
              ))
            )}
          </Flex>
        </Tabs.Panel>
        <Tabs.Panel value="comments">
          <Comments storyId={storyId} />
          <Flex align="start">
            <Box className="z-[1] flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-300">
              <Avatar
                name={session?.user?.name ?? undefined}
                size="xs"
                src={session?.user?.image ?? undefined}
              />
            </Box>
            <CommentInput storyId={storyId} />
          </Flex>
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
