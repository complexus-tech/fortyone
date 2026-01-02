"use client";
import { Flex, Skeleton } from "ui";
import { parseAsString, useQueryState } from "nuqs";
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import { LayoutSwitcher, NewStoryButton } from "@/components/ui";

export const HeaderSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const [tab] = useQueryState("tab", parseAsString.withDefault("overview"));

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <Flex align="center" gap={2}>
          <Skeleton className="size-6 rounded-full" />
          <Skeleton className="h-5 w-20 rounded" />
          <span className="hidden text-gray-300 dark:text-dark-100 md:inline">
            /
          </span>
          <Skeleton className="hidden h-5 w-24 rounded md:block" />
          <span className="hidden text-gray-300 dark:text-dark-100 md:inline">
            /
          </span>
          <Skeleton className="hidden h-5 w-16 rounded md:block" />
        </Flex>
      </Flex>
      <Flex align="center" gap={2}>
        {tab === "stories" && (
          <>
            <LayoutSwitcher layout={layout} setLayout={() => {}} />
            <Skeleton className="size-9 rounded md:size-8" />
            <Skeleton className="size-9 rounded md:size-8" />
            <span className="hidden text-text-secondary md:inline">
              |
            </span>
          </>
        )}
        <NewStoryButton className="hidden md:flex" />
      </Flex>
    </HeaderContainer>
  );
};
