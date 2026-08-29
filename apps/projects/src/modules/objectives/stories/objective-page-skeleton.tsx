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
    <Box className="flex h-full min-h-0 flex-col overflow-hidden">
      <HeaderSkeleton layout={layout} />
      <Tabs
        className="flex min-h-0 flex-1 flex-col overflow-hidden"
        value={tab as string}
      >
        <Box className="border-border d sticky top-0 z-10 flex h-[3.7rem] w-full shrink-0 items-center border-b-[0.5px] pr-12">
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
        </Box>
        <Tabs.Panel className="min-h-0 flex-1 overflow-hidden" value="overview">
          <Box className="h-full min-h-0 md:hidden">
            <OverviewSkeleton />
          </Box>
          <Box className="hidden h-full min-h-0 md:block">
            <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
              <BoardDividedPanel.MainPanel>
                <OverviewSkeleton />
              </BoardDividedPanel.MainPanel>
              <BoardDividedPanel.SideBar className="h-full min-h-0" isExpanded>
                <SidebarSkeleton />
              </BoardDividedPanel.SideBar>
            </BoardDividedPanel>
          </Box>
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1 overflow-hidden" value="stories">
          <StoriesSkeleton className="h-full" layout={layout} />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
