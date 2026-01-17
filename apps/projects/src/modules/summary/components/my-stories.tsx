"use client";
import { Avatar, Box, Button, Flex, Tabs, Text, Tooltip, Wrapper } from "ui";
import { ArrowRightIcon, CalendarIcon, StoryIcon } from "icons";
import { format, addDays, formatISO } from "date-fns";
import Link from "next/link";
import { cn } from "lib";
import { RowWrapper, PriorityIcon, StoryStatusIcon } from "@/components/ui";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import type { Story } from "@/modules/stories/types";
import { slugify } from "@/utils";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { MyStoriesSkeleton } from "./my-stories-skeleton";
import { useState } from "react";

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
  const { withWorkspace } = useWorkspacePath();
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
    <Link href={withWorkspace(`/story/${id}/${slugify(title)}`)}>
      <RowWrapper className="gap-4 px-0 md:px-0" key={id}>
        <Flex align="center" className="relative select-none" gap={2}>
          <Flex align="center" gap={2}>
            <Text className="hidden opacity-80 md:block" color="muted">
              {getTeamLabel()}-{sequenceId}
            </Text>
            <PriorityIcon className="relative -top-px" priority={priority} />
            <Text className="line-clamp-1 hover:opacity-90">{title}</Text>
          </Flex>
        </Flex>

        <Flex align="center" className="shrink-0" gap={3}>
          <Text className="flex shrink-0 items-center gap-1">
            <StoryStatusIcon className="relative -top-px" statusId={statusId} />
            <span className="hidden max-w-[16ch] truncate md:inline-block">
              {getStoryStatus()}
            </span>
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
    <Box className="border-border mt-2.5 border-t-[0.5px]">
      {stories.slice(0, 9).map((story) => (
        <StoryRow key={story.id} {...story} />
      ))}
    </Box>
  );
};

export const MyStories = () => {
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const [activeTab, setActiveTab] = useState("inProgress");

  const now = formatISO(new Date(), { representation: "date" });
  const sevenDaysLater = formatISO(addDays(new Date(), 7), {
    representation: "date",
  });

  const { data: inProgressGrouped, isPending: isInProgressPending } =
    useMyStoriesGrouped("none", {
      categories: ["started"],
      assignedToMe: true,
      storiesPerGroup: 9,
    });

  const { data: upcomingGrouped, isPending: isUpcomingPending } =
    useMyStoriesGrouped("none", {
      assignedToMe: true,
      categories: ["backlog", "unstarted", "started"],
      deadlineAfter: now,
      deadlineBefore: sevenDaysLater,
      storiesPerGroup: 9,
    });

  const { data: overdueGrouped, isPending: isOverduePending } =
    useMyStoriesGrouped("none", {
      assignedToMe: true,
      categories: ["backlog", "unstarted", "started"],
      deadlineBefore: now,
      storiesPerGroup: 9,
    });

  // Extract stories from grouped data structure
  const getStoriesFromGrouped = (
    grouped: typeof inProgressGrouped,
  ): Story[] => {
    if (!grouped) return [];
    return grouped.groups.flatMap((group) => group.stories);
  };

  const inProgressStories = getStoriesFromGrouped(inProgressGrouped);
  const upcomingStories = getStoriesFromGrouped(upcomingGrouped);
  const overdueStories = getStoriesFromGrouped(overdueGrouped);

  // Determine loading state based on active tab
  const isPending =
    (activeTab === "inProgress" && isInProgressPending) ||
    (activeTab === "upcoming" && isUpcomingPending) ||
    (activeTab === "due" && isOverduePending);

  if (isPending) {
    return <MyStoriesSkeleton />;
  }

  return (
    <Wrapper className="min-h-100 md:min-h-120">
      <Flex align="center" className="mb-2 md:mb-0" justify="between">
        <Text className="mb-2" fontSize="lg">
          Recent {getTermDisplay("storyTerm", { variant: "plural" })}
        </Text>
        <Button
          color="tertiary"
          href={withWorkspace("/my-work")}
          rightIcon={<ArrowRightIcon className="h-4" />}
          size="sm"
        >
          More {getTermDisplay("storyTerm", { variant: "plural" })}
        </Button>
      </Flex>
      <Tabs onValueChange={setActiveTab} value={activeTab}>
        <Tabs.List className="mx-0 mb-2 md:mx-0 md:mb-0">
          <Tabs.Tab value="inProgress">In Progress</Tabs.Tab>
          <Tabs.Tab value="upcoming">Due soon</Tabs.Tab>
          <Tabs.Tab value="due">Overdue</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="inProgress">
          {inProgressStories.length === 0 ? (
            <Flex
              align="center"
              className="h-100"
              direction="column"
              gap={3}
              justify="center"
            >
              <StoryIcon className="h-24 opacity-70" />
              <Text color="muted">
                You do not have any{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })} in
                progress.
              </Text>
            </Flex>
          ) : (
            <List stories={inProgressStories} />
          )}
        </Tabs.Panel>
        <Tabs.Panel value="upcoming">
          {upcomingStories.length === 0 ? (
            <Flex
              align="center"
              className="h-100"
              direction="column"
              gap={3}
              justify="center"
            >
              <StoryIcon className="h-24 opacity-70" />
              <Text color="muted">
                You do not have any{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })} due soon.
              </Text>
            </Flex>
          ) : (
            <List stories={upcomingStories} />
          )}
        </Tabs.Panel>
        <Tabs.Panel value="due">
          {overdueStories.length === 0 ? (
            <Flex
              align="center"
              className="h-100"
              direction="column"
              gap={3}
              justify="center"
            >
              <StoryIcon className="h-24 opacity-70" />
              <Text color="muted">
                You do not have any{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })} due.
              </Text>
            </Flex>
          ) : (
            <List stories={overdueStories} />
          )}
        </Tabs.Panel>
      </Tabs>
    </Wrapper>
  );
};
