"use client";
import { Box, Tabs } from "ui";
import { useSession } from "next-auth/react";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useMemo } from "react";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { MyStoriesSkeleton } from "@/modules/summary/components/my-stories-skeleton";
import { useMyStories } from "../hooks/my-stories";
import { useMyWork } from "./provider";

export const ListMyWork = ({ layout }: { layout: StoriesLayout }) => {
  const { getTermDisplay } = useTerminology();
  const { viewOptions } = useMyWork();
  const { data } = useSession();

  const user = data?.user;
  const { data: stories = [], isPending } = useMyStories();
  const tabs = ["all", "assigned", "created"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );

  const filteredStories = useMemo(() => {
    if (tab === "assigned")
      return stories.filter((story) => story.assigneeId === user?.id);
    if (tab === "created")
      return stories.filter((story) => story.reporterId === user?.id);
    return stories;
  }, [stories, tab, user?.id]);

  if (isPending) {
    return <MyStoriesSkeleton />;
  }

  return (
    <Box className="h-[calc(100vh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as typeof tab)} value={tab}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 dark:border-dark-100">
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
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={filteredStories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={filteredStories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={filteredStories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
