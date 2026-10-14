"use client";
import { Flex, Skeleton, Tabs } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { parseAsString, useQueryState } from "nuqs";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import { LayoutSwitcher } from "@/components/ui";
import { useTerminology } from "@/hooks";

export const HeaderSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const { getTermDisplay } = useTerminology();
  const [tab] = useQueryState("tab", parseAsString.withDefault("overview"));

  return (
    <HeaderContainer className="h-auto min-h-(--app-page-header-height) flex-wrap justify-between gap-x-4 gap-y-2 py-3">
      <Flex className="min-w-0 flex-1 basis-48" gap={2}>
        <Flex align="center" gap={2}>
          <Skeleton className="size-6 rounded-full" />
          <Skeleton className="h-5 w-20 rounded" />
          <span className="text-text-muted hidden md:inline">/</span>
          <Skeleton className="hidden h-5 w-24 rounded md:block" />
        </Flex>
      </Flex>
      <Flex
        align="center"
        className="ml-auto max-w-full flex-wrap justify-end"
        gap={2}
      >
        <Tabs.List
          aria-label="Objective sections"
          className="mx-0 h-min shrink-0 flex-nowrap md:mx-0"
        >
          <Tabs.Tab
            disabled
            leftIcon={<ObjectiveIcon className="size-4" />}
            value="overview"
          >
            Overview
          </Tabs.Tab>
          <Tabs.Tab
            disabled
            leftIcon={<StoryIcon className="size-4" />}
            value="stories"
          >
            {getTermDisplay("storyTerm", {
              variant: "plural",
              capitalize: true,
            })}
          </Tabs.Tab>
        </Tabs.List>
        {tab === "stories" && (
          <>
            <LayoutSwitcher layout={layout} setLayout={() => {}} />
            <Skeleton className="size-9 rounded md:size-8" />
            <Skeleton className="size-9 rounded md:size-8" />
          </>
        )}
        <Flex align="center" gap={1}>
          <Skeleton className="size-8 rounded" />
          <Skeleton className="size-8 rounded" />
          <Skeleton className="size-8 rounded" />
        </Flex>
      </Flex>
    </HeaderContainer>
  );
};
