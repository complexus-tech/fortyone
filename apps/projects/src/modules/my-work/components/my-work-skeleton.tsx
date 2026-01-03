"use client";
import { Box, Tabs } from "ui";
import { parseAsString, useQueryState } from "nuqs";
import { cn } from "lib";
import type { StoriesLayout } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";

export const MyWorkSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const [tab] = useQueryState("tab", parseAsString.withDefault("all"));
  const { getTermDisplay } = useTerminology();
  return (
    <Box className="h-[calc(100dvh-4rem)]">
      <Tabs defaultValue={tab}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-border">
          <Tabs.List>
            <Tabs.Tab value="all">
              All {getTermDisplay("storyTerm", { variant: "plural" })}
            </Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
      </Tabs>
      <BoardSkeleton
        className={cn({
          "h-[calc(100dvh-7.7rem)]": layout === "kanban",
        })}
        layout={layout}
      />
    </Box>
  );
};
