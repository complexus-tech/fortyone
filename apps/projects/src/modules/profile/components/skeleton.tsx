"use client";
import { Box, Tabs } from "ui";
import { parseAsString, useQueryState } from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { BoardSkeleton } from "@/components/ui/board-skeleton";

export const Skeleton = ({ layout }: { layout: StoriesLayout }) => {
  const [tab] = useQueryState("tab", parseAsString.withDefault("assigned"));
  return (
    <Box className="h-(--app-page-content-height) flex min-h-0 flex-col">
      <Tabs className="shrink-0" defaultValue={tab}>
        <Box className="border-border sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px]">
          <Tabs.List>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
      </Tabs>
      <BoardSkeleton className="h-auto min-h-0 flex-1" layout={layout} />
    </Box>
  );
};
