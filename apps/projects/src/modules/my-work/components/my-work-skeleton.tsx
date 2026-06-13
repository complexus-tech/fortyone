"use client";
import { Box, Container, Skeleton, Tabs } from "ui";
import { parseAsString, useQueryState } from "nuqs";
import { cn } from "lib";
import type { StoriesLayout } from "@/components/ui";
import { useTerminology, useUserRole } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";

export const MyWorkSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  const { userRole } = useUserRole();
  const isAdmin = userRole === "admin";
  const [tab] = useQueryState(
    "tab",
    parseAsString.withDefault(isAdmin ? "pulse" : "all"),
  );
  const { getTermDisplay } = useTerminology();
  const isPulseTab = isAdmin && tab === "pulse";

  return (
    <Box className="h-[calc(100dvh-4rem)]">
      <Tabs defaultValue={tab}>
        <Box className="border-border sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px]">
          <Tabs.List>
            {isAdmin ? <Tabs.Tab value="pulse">Pulse</Tabs.Tab> : null}
            <Tabs.Tab value="all">
              All {getTermDisplay("storyTerm", { variant: "plural" })}
            </Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
      </Tabs>
      {isPulseTab ? (
        <Container className="@container py-4">
          <Skeleton className="mb-3 h-8 w-72" />
          <Skeleton className="mb-5 h-5 w-full max-w-xl" />
          <Box className="mb-4 grid grid-cols-2 gap-3 @3xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton className="h-28" key={index} />
            ))}
          </Box>
          <Box className="grid gap-4 @5xl:grid-cols-2">
            <Skeleton className="h-80" />
            <Skeleton className="h-80" />
          </Box>
        </Container>
      ) : (
        <BoardSkeleton
          className={cn({
            "h-[calc(100dvh-7.7rem)]": layout === "kanban",
          })}
          layout={layout}
        />
      )}
    </Box>
  );
};
