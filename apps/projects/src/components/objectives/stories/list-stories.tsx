"use client";
import { BreadCrumbs, Flex, Box, Tabs } from "ui";
import { StoryIcon } from "icons";
import type { Story, StoryStatus } from "@/types/story";
import { useLocalStorage } from "@/hooks";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesBoard,
  LayoutSwitcher,
  NewStoryButton,
  SideDetailsSwitch,
  BoardDividedPanel,
  StoriesFiltersButton,
} from "@/components/ui";
import { Sidebar } from "./sidebar";

export const ListStories = ({
  stories,
  statuses,
}: {
  stories: Story[];
  statuses: StoryStatus[];
}) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:stories:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "objective:stories:expanded",
    true,
  );

  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Engineering",
              icon: "🚀",
            },
            {
              name: "Stories",
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={setLayout} />
          <StoriesFiltersButton />
          <NewStoryButton />
          <span className="text-gray-200 dark:text-dark-100">|</span>
          <SideDetailsSwitch
            isExpanded={isExpanded}
            setIsExpanded={setIsExpanded}
          />
        </Flex>
      </HeaderContainer>

      <BoardDividedPanel autoSaveId="objective:stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <Tabs defaultValue="all">
            <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-gray-100/60 backdrop-blur-lg dark:border-dark-100">
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
                statuses={statuses}
                stories={stories}
              />
            </Tabs.Panel>
            <Tabs.Panel value="active">
              <StoriesBoard
                className="h-[calc(100vh-7.7rem)]"
                layout={layout}
                statuses={statuses}
                stories={stories}
              />
            </Tabs.Panel>
            <Tabs.Panel value="backlog">
              <StoriesBoard
                className="h-[calc(100vh-7.7rem)]"
                layout={layout}
                statuses={statuses}
                stories={stories}
              />
            </Tabs.Panel>
          </Tabs>
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
