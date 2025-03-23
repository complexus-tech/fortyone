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
          <span className="text-gray-300 dark:text-dark-100">/</span>
          <Skeleton className="h-5 w-24 rounded" />
          <span className="text-gray-300 dark:text-dark-100">/</span>
          <Skeleton className="h-5 w-16 rounded" />
        </Flex>
      </Flex>
      <Flex align="center" gap={2}>
        {tab === "stories" && (
          <>
            <LayoutSwitcher layout={layout} setLayout={() => {}} />
            <Skeleton className="h-8 w-8 rounded" />
            <Skeleton className="h-8 w-8 rounded" />
            <span className="text-gray-200 dark:text-dark-100">|</span>
          </>
        )}
        <NewStoryButton />
      </Flex>
    </HeaderContainer>
  );
};
