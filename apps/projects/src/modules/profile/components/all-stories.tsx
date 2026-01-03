"use client";
import { Box, Tabs } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/modules/stories/types";
import { useGroupedStories } from "@/modules/stories/hooks/use-grouped-stories";
import { useProfile } from "./provider";
import { Skeleton } from "./skeleton";

export const AllStories = ({
  layout,
}: {
  stories: Story[];
  layout: StoriesLayout;
}) => {
  const { userId } = useParams<{
    userId: string;
  }>();
  const tabs = ["assigned", "created"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("assigned"),
  );
  const { viewOptions } = useProfile();

  const { data: groupedStories, isPending } = useGroupedStories({
    groupBy: viewOptions.groupBy,
    assigneeIds: tab === "assigned" ? [userId] : undefined,
    reporterIds: tab === "created" ? [userId] : undefined,
    orderBy: viewOptions.orderBy,
  });

  if (isPending) return <Skeleton layout={layout} />;

  return (
    <Box className="h-[calc(100vh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as typeof tab)} value={tab}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-border backdrop-blur-lg d/40">
          <Tabs.List>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="assigned">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesBoard
            className="h-[calc(100vh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
