"use client";
import { Box, Tabs } from "ui";
import { useSession } from "next-auth/react";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { useMyStories } from "../hooks/my-stories";
import { useMyWork } from "./provider";

export const ListMyWork = ({ layout }: { layout: StoriesLayout }) => {
  const { viewOptions } = useMyWork();
  const { data } = useSession();
  const user = data?.user;
  const { data: stories = [] } = useMyStories();
  const tabs = ["assigned", "created", "subscribed"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("assigned"),
  );

  const assinedStories = stories.filter(
    (story) => story.assigneeId === user?.id,
  );
  const createdStories = stories.filter(
    (story) => story.reporterId === user?.id,
  );

  return (
    <Box className="h-[calc(100vh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as typeof tab)} value={tab}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 dark:border-dark-100">
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
            stories={assinedStories}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            layout={layout}
            stories={createdStories}
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
