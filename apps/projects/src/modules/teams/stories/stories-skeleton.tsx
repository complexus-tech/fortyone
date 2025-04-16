"use client";
import { Box, Tabs, BreadCrumbs, Flex, Skeleton } from "ui";
import { parseAsString, useQueryState } from "nuqs";
import { cn } from "lib";
import type { StoriesLayout } from "@/components/ui";
import { LayoutSwitcher } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";

export const StoriesSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const [tab] = useQueryState("tab", parseAsString.withDefault("all"));
  const { getTermDisplay } = useTerminology();
  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Team",
                icon: <Skeleton className="size-4 rounded" />,
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={() => {}} />
          <Skeleton className="h-9 w-9 md:h-6 md:w-16" />
          <Skeleton className="hidden h-6 w-20 md:block" />
          <Skeleton className="hidden h-6 w-16 md:block" />
        </Flex>
      </HeaderContainer>
      <Box className="2xl:h-[calc(100dvh-4rem)]">
        <Tabs defaultValue={tab}>
          <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 dark:border-dark-100">
            <Tabs.List>
              <Tabs.Tab value="all">
                All {getTermDisplay("storyTerm", { variant: "plural" })}
              </Tabs.Tab>
              <Tabs.Tab value="active">Active</Tabs.Tab>
              <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
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
    </>
  );
};
