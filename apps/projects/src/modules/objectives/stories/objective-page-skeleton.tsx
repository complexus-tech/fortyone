"use client";
import { Box, Tabs } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { HeaderSkeleton } from "./header-skeleton";
import { OverviewSkeleton } from "./overview/overview-skeleton";
import { StoriesSkeleton } from "./stories-skeleton";
import { SidebarSkeleton } from "./sidebar-skeleton";

export const ObjectivePageSkeleton = ({
  layout,
}: {
  layout: StoriesLayout;
}) => {
  const { getTermDisplay } = useTerminology();
  const tabs = ["overview", "stories"] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("overview"),
  );

  return (
    <>
      <HeaderSkeleton layout={layout} />
      <Tabs value={tab as string}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full items-center justify-between border-b-[0.5px] border-gray-100/60 pr-12 dark:border-dark-100">
          <Tabs.List className="h-min">
            <Tabs.Tab leftIcon={<ObjectiveIcon />} value="overview">
              Overview
            </Tabs.Tab>
            <Tabs.Tab leftIcon={<StoryIcon />} value="stories">
              {getTermDisplay("storyTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Tabs.Tab>
          </Tabs.List>
          <Box className="h-8 w-24">
            {/* Placeholder for the ordering or copy button */}
          </Box>
        </Box>
        <Tabs.Panel value="overview">
          <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
            <BoardDividedPanel.MainPanel>
              <OverviewSkeleton />
            </BoardDividedPanel.MainPanel>
            <BoardDividedPanel.SideBar
              className="h-[calc(100vh-7.7rem)]"
              isExpanded
            >
              <SidebarSkeleton />
            </BoardDividedPanel.SideBar>
          </BoardDividedPanel>
        </Tabs.Panel>
        <Tabs.Panel value="stories">
          <StoriesSkeleton layout={layout} />
        </Tabs.Panel>
      </Tabs>
    </>
  );
};
