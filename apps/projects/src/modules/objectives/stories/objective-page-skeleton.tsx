"use client";
import { Box, Tabs } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { HeaderSkeleton } from "./header-skeleton";
import { OverviewSkeleton } from "./overview/overview-skeleton";
import { StoriesSkeleton } from "./stories-skeleton";
import { SidebarSkeleton } from "./sidebar-skeleton";

export const ObjectivePageSkeleton = ({
  layout,
}: {
  layout: StoriesLayout;
}) => {
  const tabs = ["overview", "stories"] as const;
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("overview"),
  );

  return (
    <Box className="flex h-full min-h-0 flex-col overflow-hidden">
      <Tabs
        className="flex min-h-0 flex-1 flex-col overflow-hidden"
        value={tab as string}
      >
        <HeaderSkeleton layout={layout} />
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
