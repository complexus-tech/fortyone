"use client";
import { Box, Tabs } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { MyWorkSkeleton } from "@/modules/my-work/components/my-work-skeleton";
import { useMyWork } from "./provider";

export const ListMyWork = ({ layout }: { layout: StoriesLayout }) => {
  const tabs = ["all", "assigned", "created"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  const { getTermDisplay } = useTerminology();
  const { viewOptions } = useMyWork();

  const { data: groupedStories, isPending } = useMyStoriesGrouped(
    viewOptions.groupBy,
    {
      createdByMe: tab === "created" || tab === "all" ? true : undefined,
      assignedToMe: tab === "assigned" || tab === "all" ? true : undefined,
    },
  );

  if (isPending) return <MyWorkSkeleton layout={layout} />;

  return (
    <Box className="h-[calc(100dvh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as typeof tab)} value={tab}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100 dark:border-dark-100">
          <Tabs.List>
            <Tabs.Tab value="all">
              All {getTermDisplay("storyTerm", { variant: "plural" })}
            </Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="all">
          <StoriesBoard
            className="h-[calc(100dvh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <StoriesBoard
            className="h-[calc(100dvh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesBoard
            className="h-[calc(100dvh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
