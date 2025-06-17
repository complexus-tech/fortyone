"use client";
import { Flex, Text, ProgressBar, Box, Badge, Tooltip } from "ui";
import Link from "next/link";
import { ArrowRightIcon, CalendarIcon, SprintsIcon } from "icons";
import { format } from "date-fns";
import { QueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { RowWrapper } from "@/components/ui/row-wrapper";
import type { Sprint } from "@/modules/sprints/types";
import { StoryStatusIcon } from "@/components/ui";
import { storyKeys } from "@/modules/stories/constants";
import { getStories } from "@/modules/stories/queries/get-stories";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

type SprintStatus = "completed" | "in progress" | "upcoming";

const statusColors = {
  completed: "tertiary",
  "in progress": "primary",
  upcoming: "tertiary",
} as const;

export const SprintRow = ({
  id,
  name,
  teamId,
  startDate,
  endDate,
  stats: { total, completed, started, unstarted, backlog },
}: Sprint) => {
  const { data: session } = useSession();
  const queryClient = new QueryClient();
  const startDateObj = new Date(startDate);
  const endDateObj = new Date(endDate);

  let sprintStatus: SprintStatus = "completed";
  if (startDateObj < new Date() && endDateObj > new Date()) {
    sprintStatus = "in progress";
  } else if (startDateObj > new Date()) {
    sprintStatus = "upcoming";
  }

  const progress = Math.round((completed / total) * 100) || 0;

  return (
    <RowWrapper className="gap-3">
      <Link
        className="flex flex-1 items-center gap-4"
        href={`/teams/${teamId}/sprints/${id}/stories`}
        onMouseEnter={() => {
          queryClient.prefetchQuery({
            queryKey: storyKeys.sprint(id),
            queryFn: () => getStories(session!, { sprintId: id }),
            staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 3,
          });
        }}
        prefetch
      >
        <Flex
          align="center"
          className="size-10 rounded-[0.6rem] bg-gray-100/50 dark:bg-dark-200"
          justify="center"
        >
          <SprintsIcon />
        </Flex>
        <Box className="space-y-1">
          <Text
            className="line-clamp-1 antialiased md:text-[1.05rem]"
            fontWeight="semibold"
          >
            {name}
          </Text>
          <Text className="flex items-center gap-1.5" color="muted">
            <CalendarIcon className="h-[1.1rem]" />
            {format(startDateObj, "MMM d")}
            <ArrowRightIcon className="h-3" />
            {format(endDateObj, "MMM d")}
          </Text>
        </Box>
      </Link>

      <Flex className="items-center" gap={4}>
        <Badge
          className="h-8 px-2 text-base capitalize tracking-wide"
          color={statusColors[sprintStatus]}
        >
          {sprintStatus}
        </Badge>

        <Tooltip title={`${progress}% Complete`}>
          <Flex align="center" className="hidden w-36 md:flex" gap={3}>
            <ProgressBar className="h-2 flex-1" progress={progress} />
            <Text>{progress}%</Text>
          </Flex>
        </Tooltip>

        <Flex className="hidden min-w-[300px] md:flex" gap={4}>
          <Flex align="center" className="min-w-[80px] gap-1.5">
            <StoryStatusIcon category="completed" />
            <Text>
              {completed}
              <span className="text-muted ml-1.5">Done</span>
            </Text>
          </Flex>
          <Flex align="center" className="min-w-[80px] gap-1.5">
            <StoryStatusIcon category="started" />
            <Text className="whitespace-nowrap">
              {started}
              <span className="text-muted ml-2">Active</span>
            </Text>
          </Flex>
          <Flex align="center" className="min-w-[80px] gap-1.5">
            <StoryStatusIcon category="unstarted" />
            <Text className="whitespace-nowrap">
              {unstarted}
              <span className="text-muted ml-1.5">Todo</span>
            </Text>
          </Flex>
          <Flex align="center" className="min-w-[80px] gap-1.5">
            <StoryStatusIcon category="backlog" />
            <Text className="whitespace-nowrap">
              {backlog}
              <span className="text-muted ml-1.5">Backlog</span>
            </Text>
          </Flex>
        </Flex>
      </Flex>
    </RowWrapper>
  );
};
