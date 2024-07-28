"use client";
import { Box, Tabs } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/modules/stories/types";
import { useTeamStories } from "@/modules/teams/stories/provider";
import { useStore } from "@/hooks/store";

export const AllStories = ({
  layout,
  stories,
}: {
  stories: Story[];
  layout: StoriesLayout;
}) => {
  const { states } = useStore();
  const { viewOptions } = useTeamStories();
  const backlogStatuses = states
    .filter((state) => state.category === "backlog")
    .map((state) => state.id);
  const activeStatuses = states
    .filter((state) => (state.category = "started"))
    .map((state) => state.id);

  const backlog = stories.filter((story) =>
    backlogStatuses.includes(story.statusId),
  );
  const activeStories = stories.filter((story) =>
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
          stories={stories}
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
