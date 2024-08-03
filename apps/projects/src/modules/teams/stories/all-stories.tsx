"use client";
import { Box, Tabs } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { useTeamOptions } from "@/modules/teams/stories/provider";
import { useStore } from "@/hooks/store";
import { useMemo } from "react";
import { isAfter, isBefore, isThisWeek, isToday } from "date-fns";
import { useSession } from "next-auth/react";
import { useParams } from "next/navigation";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: stories = [] } = useTeamStories(teamId);
  const session = useSession();
  const { states, sprints } = useStore();
  // a sprint is active if it has a start date and end date and the current date is between the start and end date
  // use date-fns to check if the current date is between the start and end date
  const activeSprints = sprints
    .filter(
      (sprint) =>
        sprint.startDate &&
        sprint.endDate &&
        isAfter(new Date(), new Date(sprint.startDate)) &&
        isBefore(new Date(), new Date(sprint.endDate)),
    )
    .map((sprint) => sprint.id);

  const completedStatuses = states
    .filter((state) => state.category === "completed")
    .map((state) => state.id);
  const { viewOptions, filters } = useTeamOptions();
  const backlogStatuses = states
    .filter((state) => state.category === "backlog")
    .map((state) => state.id);
  const activeStatuses = states
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
    <Tabs defaultValue="all">
      <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 dark:border-dark-100">
        <Tabs.List>
          <Tabs.Tab value="all">All stories</Tabs.Tab>
          <Tabs.Tab value="active">Active</Tabs.Tab>
          <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
        </Tabs.List>
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
