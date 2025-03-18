/* eslint-disable react/no-array-index-key -- ok for skeleton screens */
"use client";
import { Box, Checkbox, Flex, Skeleton, Tabs } from "ui";
import { parseAsString, useQueryState } from "nuqs";
import { cn } from "lib";
import type { StoriesLayout } from "@/components/ui";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { useTerminology } from "@/hooks";

export const MyWorkSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const [tab] = useQueryState("tab", parseAsString.withDefault("all"));
  const { getTermDisplay } = useTerminology();
  return (
    <Box className="2xl:h-[calc(100vh-4rem)]">
      <Tabs defaultValue={tab}>
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 dark:border-dark-100">
          <Tabs.List>
            <Tabs.Tab value="all">
              All {getTermDisplay("storyTerm", { variant: "plural" })}
            </Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
      </Tabs>

      {/* Content skeleton based on layout */}
      {layout === "kanban" ? <KanbanLayoutSkeleton /> : <ListLayoutSkeleton />}
    </Box>
  );
};

const KanbanLayoutSkeleton = () => {
  return (
    <Box className="overflow-x-auto bg-gray-50/60 dark:bg-transparent 2xl:h-[calc(100vh-7.7rem)]">
      {/* Kanban header */}
      <Box className="sticky top-0 z-[1] h-[3.5rem] w-max px-6 backdrop-blur">
        <Flex align="center" className="h-full shrink-0" gap={6}>
          {Array.from({ length: 4 }).map((_, i) => (
            <Flex
              align="center"
              className="w-[280px]"
              justify="between"
              key={i}
            >
              <Flex align="center" gap={2}>
                <Skeleton className="h-5 w-5 rounded-full" />
                <Skeleton className="h-5 w-24" />
              </Flex>
              <Skeleton className="size-6 rounded-full" />
            </Flex>
          ))}
        </Flex>
      </Box>

      {/* Kanban columns */}
      <Box className="flex w-max gap-x-6 px-7 2xl:h-[calc(100%-3.5rem)]">
        {Array.from({ length: 4 }).map((_, i) => (
          <Box className="w-[280px]" key={i}>
            {Array.from({ length: 6 - i }).map((_, j) => (
              <Skeleton
                className="mb-3 h-28 shadow-sm dark:shadow-none"
                key={j}
              />
            ))}
          </Box>
        ))}
      </Box>
    </Box>
  );
};

const ListLayoutSkeleton = () => {
  return (
    <Box className="overflow-x-auto pb-6 2xl:h-[calc(100vh-7.7rem)]">
      {Array.from({ length: 8 }).map((_, i) => (
        <RowWrapper className="relative gap-4" key={i}>
          <Checkbox className="absolute left-5 opacity-70" />
          <Flex align="center" className="relative shrink" gap={3}>
            <Skeleton className="h-5 w-10" />
            <Skeleton
              className={cn("h-5 w-32", {
                "w-56": i % 2 === 0,
              })}
            />
          </Flex>
          <Flex align="center" className="shrink-0" gap={4}>
            <Skeleton className="h-5 w-20" />
            <Skeleton className="h-5 w-16" />
            <Skeleton className="h-5 w-20" />
            <Skeleton className="size-8 rounded-full" />
          </Flex>
        </RowWrapper>
      ))}
    </Box>
  );
};
