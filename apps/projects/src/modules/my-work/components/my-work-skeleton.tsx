"use client";
import { Box, Flex, Skeleton, Tabs } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { HeaderContainer } from "@/components/shared";

const tabs = ["all", "assigned", "collaborating", "created"] as const;

export const MyWorkSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const [tab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  const { getTermDisplay } = useTerminology();

  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex align="center">
          <Skeleton className="h-6 w-40" />
        </Flex>
        <Flex align="center" gap={2}>
          <Tabs defaultValue={tab}>
            <Tabs.List className="mx-0 flex-nowrap md:mx-0">
              <Tabs.Tab value="all">
                All {getTermDisplay("storyTerm", { variant: "plural" })}
              </Tabs.Tab>
              <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
              <Tabs.Tab value="collaborating">Collaborating</Tabs.Tab>
              <Tabs.Tab value="created">Created</Tabs.Tab>
            </Tabs.List>
          </Tabs>
          <span
            aria-hidden="true"
            className="bg-border mx-1 h-5 w-px shrink-0"
          />
          <Skeleton className="h-8 w-20 rounded-xl" />
          <Skeleton className="h-8 w-16 rounded-xl" />
          <Skeleton className="h-8 w-24 rounded-xl" />
        </Flex>
      </HeaderContainer>
      <Box className="h-[calc(100%-3.6rem)]">
        <BoardSkeleton className="h-full" layout={layout} />
      </Box>
    </>
  );
};
