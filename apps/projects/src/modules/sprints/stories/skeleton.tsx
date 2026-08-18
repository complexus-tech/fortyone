"use client";
import { BreadCrumbs, Flex, Skeleton } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { LayoutSwitcher } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { HeaderContainer } from "@/components/shared";

export const StoriesSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const { getTermDisplay } = useTerminology();
  return (
    <>
      <HeaderContainer className="justify-between">
        <Flex gap={2}>
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Team",
              },
              {
                name: getTermDisplay("sprintTerm", {
                  variant: "plural",
                  capitalize: true,
                }),
              },
              {
                name: getTermDisplay("storyTerm", {
                  variant: "plural",
                  capitalize: true,
                }),
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={() => {}} />
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-6 w-16" />
        </Flex>
      </HeaderContainer>
      <BoardSkeleton layout={layout} />
    </>
  );
};
