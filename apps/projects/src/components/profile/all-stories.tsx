"use client";
import { Box, Tabs } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/types/story";
import { useProfile } from "./provider";

export const AllStories = ({
  stories,
  layout,
}: {
  stories: Story[];
  layout: StoriesLayout;
}) => {
  const { viewOptions } = useProfile();
  return (
    <Box className="h-[calc(100vh-4rem)]">
      <Tabs defaultValue="assigned">
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-gray-100 backdrop-blur-lg dark:border-dark-100/40">
          <Tabs.List>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
            <Tabs.Tab value="subscribed">Subscribed</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="assigned">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={stories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={stories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="subscribed">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={stories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
