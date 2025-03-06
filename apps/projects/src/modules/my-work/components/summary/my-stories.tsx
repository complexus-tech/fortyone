"use client";
import { Avatar, Box, Button, Flex, Tabs, Text, Tooltip, Wrapper } from "ui";
import { ArrowRightIcon, CalendarIcon, StoryIcon } from "icons";
import { useSession } from "next-auth/react";
import { format, addDays } from "date-fns";
import Link from "next/link";
import { cn } from "lib";
import { RowWrapper, PriorityIcon, StoryStatusIcon } from "@/components/ui";
import { useMyStories } from "@/modules/my-work/hooks/my-stories";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import type { Story } from "@/modules/stories/types";
import { slugify } from "@/utils";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";

const StoryRow = ({
  id,
  title,
  priority,
  statusId,
  sequenceId,
  teamId,
  assigneeId,
  endDate,
}: Story) => {
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();

  const getTeamLabel = () => {
    return teams.find((team) => team.id === teamId)?.code;
  };

  const getStoryStatus = () => {
    return statuses.find((status) => status.id === statusId)?.name;
  };

  const getStoryMember = () => {
    return members.find((member) => member.id === assigneeId);
  };

  return (
    <Link href={`/story/${id}/${slugify(title)}`}>
      <RowWrapper className="gap-4 px-0" key={id}>
        <Flex align="center" className="relative select-none" gap={2}>
          <Flex align="center" gap={2}>
            <Text className="opacity-80" color="muted" fontWeight="normal">
              {getTeamLabel()}-{sequenceId}
            </Text>
            <PriorityIcon className="relative -top-px" priority={priority} />
            <Text className="max-w-[22rem] truncate hover:opacity-90">
              {title}
            </Text>
          </Flex>
        </Flex>

        <Flex align="center" gap={3}>
          <Text className="flex shrink-0 items-center gap-1">
            <StoryStatusIcon className="relative -top-px" statusId={statusId} />
            {getStoryStatus()}
          </Text>
          {endDate ? (
            <Tooltip
              title={
                <Flex align="start" gap={2}>
                  <CalendarIcon
                    className={cn("relative top-[2.5px] h-5 w-auto", {
                      "text-primary dark:text-primary":
                        new Date(endDate) < new Date(),
                      "text-warning dark:text-warning":
                        new Date(endDate) <= addDays(new Date(), 7) &&
                        new Date(endDate) >= new Date(),
                    })}
                  />
                  <Box>{getDueDateMessage(new Date(endDate))}</Box>
                </Flex>
              }
            >
              <Text
                className={cn("flex shrink-0 items-center gap-1", {
                  "text-primary dark:text-primary":
                    new Date(endDate) < new Date(),
                  "text-warning dark:text-warning":
                    new Date(endDate) <= addDays(new Date(), 7) &&
                    new Date(endDate) >= new Date(),
                })}
              >
                <CalendarIcon
                  className={cn("relative -top-px", {
                    "text-primary dark:text-primary":
                      new Date(endDate) < new Date(),
                    "text-warning dark:text-warning":
                      new Date(endDate) <= addDays(new Date(), 7) &&
                      new Date(endDate) >= new Date(),
                  })}
                />
                {format(new Date(endDate), "MMM, d")}
              </Text>
            </Tooltip>
          ) : null}
          <Avatar
            name={getStoryMember()?.fullName}
            size="xs"
            src={getStoryMember()?.avatarUrl}
          />
        </Flex>
      </RowWrapper>
    </Link>
  );
};

const List = ({ stories }: { stories: Story[] }) => {
  return (
    <Box className="mt-2.5 border-t border-gray-50 dark:border-dark-200">
      {stories.slice(0, 9).map((story) => (
        <StoryRow key={story.id} {...story} />
      ))}
    </Box>
  );
};

export const MyStories = () => {
  const { data: session } = useSession();
  const { data: stories = [] } = useMyStories();
  const { data: statuses = [] } = useStatuses();

  const inProgressStatuses = statuses
    .filter((status) => {
      return status.category === "started";
    })
    .map((status) => status.id);

  const completedOrCancelledStatuses = statuses
    .filter((status) => {
      return (
        status.category === "completed" ||
        status.category === "cancelled" ||
        status.category === "paused"
      );
    })
    .map((status) => status.id);

  const upcomingDueDates = stories.filter((story) => {
    return (
      story.endDate &&
      new Date(story.endDate) > new Date() &&
      story.assigneeId === session?.user?.id
    );
  });

  const dueStories = stories.filter((story) => {
    return (
      story.endDate &&
      new Date(story.endDate) < new Date() &&
      story.assigneeId === session?.user?.id &&
      !completedOrCancelledStatuses.includes(story.statusId)
    );
  });

  const inProgressStories = stories.filter((story) => {
    return (
      inProgressStatuses.includes(story.statusId) &&
      story.assigneeId === session?.user?.id
    );
  });

  return (
    <Wrapper>
      <Flex align="center" justify="between">
        <Text className="mb-2" fontSize="lg">
          Recent stories
        </Text>
        <Button
          color="tertiary"
          href="/my-work"
          rightIcon={<ArrowRightIcon className="h-4" />}
          size="sm"
        >
          More stories
        </Button>
      </Flex>
      <Tabs defaultValue="inProgress">
        <Tabs.List className="mx-0">
          <Tabs.Tab value="inProgress">In Progress</Tabs.Tab>
          <Tabs.Tab value="upcoming">Due soon</Tabs.Tab>
          <Tabs.Tab value="due">Overdue</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="inProgress">
          {inProgressStories.length === 0 ? (
            <Flex
              align="center"
              className="h-[25rem]"
              direction="column"
              gap={3}
              justify="center"
            >
              <StoryIcon className="h-24 opacity-70" />
              <Text color="muted">
                You do not have any stories in progress.
              </Text>
            </Flex>
          ) : (
            <List stories={inProgressStories} />
          )}
        </Tabs.Panel>
        <Tabs.Panel value="upcoming">
          {upcomingDueDates.length === 0 ? (
            <Flex
              align="center"
              className="h-[25rem]"
              direction="column"
              gap={3}
              justify="center"
            >
              <StoryIcon className="h-24 opacity-70" />
              <Text color="muted">You do not have any stories due soon.</Text>
            </Flex>
          ) : (
            <List stories={upcomingDueDates} />
          )}
        </Tabs.Panel>
        <Tabs.Panel value="due">
          {dueStories.length === 0 ? (
            <Flex
              align="center"
              className="h-[25rem]"
              direction="column"
              gap={3}
              justify="center"
            >
              <StoryIcon className="h-24 opacity-70" />
              <Text color="muted">You do not have any stories due.</Text>
            </Flex>
          ) : (
            <List stories={dueStories} />
          )}
        </Tabs.Panel>
      </Tabs>
    </Wrapper>
  );
};
