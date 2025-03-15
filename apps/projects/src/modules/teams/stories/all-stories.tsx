"use client";
import { Box, Tabs, Text } from "ui";
import { useMemo } from "react";
import { isAfter, isBefore, isThisWeek, isToday } from "date-fns";
import { useSession } from "next-auth/react";
import { useParams } from "next/navigation";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { ArrowUpDownIcon } from "icons";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useTeamOptions } from "@/modules/teams/stories/provider";
import { StoriesBoard } from "@/components/ui";
import type { StoriesLayout } from "@/components/ui";
import { useStatuses } from "@/lib/hooks/statuses";
import { useSprints } from "@/modules/sprints/hooks/sprints";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: stories = [] } = useTeamStories(teamId);
  const session = useSession();
  const { data: statuses = [] } = useStatuses();
  const { data: sprints = [] } = useSprints();

  const tabs = ["all", "active", "backlog"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  type Tab = (typeof tabs)[number];
  // a sprint is active if it has a start date and deadline and the current date is between the start and deadline
  // use date-fns to check if the current date is between the start and deadline
  const activeSprints = sprints
    .filter(
      (sprint) =>
        sprint.startDate &&
        sprint.endDate &&
        isAfter(new Date(), new Date(sprint.startDate)) &&
        isBefore(new Date(), new Date(sprint.endDate)),
    )
    .map((sprint) => sprint.id);

  const completedStatuses = statuses
    .filter((state) => state.category === "completed")
    .map((state) => state.id);
  const { viewOptions, filters } = useTeamOptions();
  const backlogStatuses = statuses
    .filter((state) => state.category === "backlog")
    .map((state) => state.id);
  const activeStatuses = statuses
    .filter((state) => state.category === "started")
    .map((state) => state.id);

  //filters
  const filteredStories = useMemo(() => {
    let newStories = [...stories];
    if (filters.completed) {
      newStories = newStories.filter((story) =>
        completedStatuses.includes(story.statusId),
      );
    }
    if (filters.dueThisWeek) {
      newStories = newStories.filter((story) =>
        isThisWeek(new Date(story.endDate!)),
      );
    }
    if (filters.dueToday) {
      newStories = newStories.filter((story) =>
        isToday(new Date(story.endDate!)),
      );
    }
    if (filters.completed) {
      newStories = newStories.filter((story) =>
        completedStatuses.includes(story.statusId),
      );
    }
    if (filters.assignedToMe) {
      newStories = newStories.filter(
        (story) => story.assigneeId === session.data?.user?.id,
      );
    }
    if (filters.activeSprints) {
      newStories = newStories.filter((story) =>
        activeSprints.includes(story.sprintId!),
      );
    }
    return newStories;
  }, [stories, filters]);

  const backlog = filteredStories.filter((story) =>
    backlogStatuses.includes(story.statusId),
  );
  const activeStories = filteredStories.filter((story) =>
    activeStatuses.includes(story.statusId),
  );

  return (
    <Tabs
      onValueChange={(v) => {
        setTab(v as Tab);
      }}
      value={tab}
    >
      <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full items-center justify-between border-b-[0.5px] border-gray-100/60 pr-12 dark:border-dark-100">
        <Tabs.List className="h-min">
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="active">Active</Tabs.Tab>
          <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
        </Tabs.List>
        <Text
          className="ml-2 flex shrink-0 items-center gap-1.5 px-1 opacity-80"
          color="muted"
        >
          <ArrowUpDownIcon className="h-4 w-auto" />
          Ordering by{" "}
          <span className="font-semibold capitalize">
            {viewOptions.orderBy}
          </span>
        </Text>
      </Box>

      <Tabs.Panel value="all">
        <StoriesBoard
          className="h-[calc(100vh-7.7rem)]"
          layout={layout}
          stories={filteredStories}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
      <Tabs.Panel value="active">
        <StoriesBoard
          className="h-[calc(100vh-7.7rem)]"
          layout={layout}
          stories={activeStories}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
      <Tabs.Panel value="backlog">
        <StoriesBoard
          className="h-[calc(100vh-7.7rem)]"
          layout={layout}
          stories={backlog}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
    </Tabs>
  );
};
