"use client";
import { BreadCrumbs, Flex, Box, Tabs } from "ui";
import { StoryIcon } from "icons";
import type { Story } from "@/types/story";
import { useLocalStorage } from "@/hooks";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  StoriesBoard,
  LayoutSwitcher,
  NewStoryButton,
  SideDetailsSwitch,
  BoardDividedPanel,
  StoriesViewOptionsButton,
} from "@/components/ui";
import { Sidebar } from "./sidebar";

export const ListStories = ({ stories }: { stories: Story[] }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:stories:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "objective:stories:expanded",
    true,
  );

  const backlog = stories.filter((story) => story.status === "Backlog");
  const activeIssues = stories.filter((story) => story.status !== "Backlog");

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
          <StoriesViewOptionsButton />
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
            <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 backdrop-blur-lg dark:border-dark-100">
              <Tabs.List>
                <Tabs.Tab value="active">Active</Tabs.Tab>
                <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
                <Tabs.Tab value="all">All stories</Tabs.Tab>
              </Tabs.List>
            </Box>
            <Tabs.Panel value="all">
              <StoriesBoard
                className="h-[calc(100vh-7.7rem)]"
                layout={layout}
                stories={stories}
              />
            </Tabs.Panel>
            <Tabs.Panel value="active">
              <StoriesBoard
                className="h-[calc(100vh-7.7rem)]"
                layout={layout}
                stories={activeIssues}
              />
            </Tabs.Panel>
            <Tabs.Panel value="backlog">
              <StoriesBoard
                className="h-[calc(100vh-7.7rem)]"
                layout={layout}
                stories={backlog}
              />
            </Tabs.Panel>
          </Tabs>
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar stories={stories} />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
