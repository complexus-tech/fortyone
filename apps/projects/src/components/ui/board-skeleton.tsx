import { cn } from "lib";
import { Box, Checkbox, Flex, Skeleton } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { RowWrapper } from "@/components/ui/row-wrapper";
import type { StoriesLayout } from "./stories-board";

const ListLayoutSkeleton = () => {
  return (
    <Box>
      {Array.from({ length: 4 }).map((_, groupIndex) => (
        <Box className="mb-2" key={groupIndex}>
          {/* Group header, similar to StoriesHeader */}
          <Box className="border-border bg-surface-muted sticky top-0 z-1 border-y-[0.5px] py-[0.4rem] backdrop-blur select-none">
            <Flex align="center" className="px-12" justify="between">
              <Flex align="center" className="relative" gap={2}>
                <Checkbox className="absolute -left-7 opacity-70" />
                <Flex align="center" gap={2}>
                  <Skeleton className="h-5 w-5 animate-none rounded-full" />
                  <Skeleton className="h-4 w-24 animate-none" />
                </Flex>
                <Skeleton className="ml-2 h-4 w-4 animate-none" />
                <Skeleton className="h-4 w-12 animate-none" />
              </Flex>
              <Skeleton className="h-8 w-8 animate-none rounded-full" />
            </Flex>
          </Box>
          {/* Group items */}
          {Array.from({ length: 4 - groupIndex }).map((_, i) => (
            <RowWrapper className="pointer-events-none relative gap-4" key={i}>
              <Checkbox className="absolute left-5 hidden opacity-70 md:block" />
              <Flex align="center" className="relative shrink" gap={3}>
                <Skeleton className="h-5 w-10" />
                <Skeleton
                  className={cn("h-5 w-24 md:w-32", {
                    "w-40 md:w-56": i % 2 === 0,
                  })}
                />
              </Flex>
              <Flex align="center" className="shrink-0" gap={4}>
                <Skeleton className="h-5 w-20" />
                <Skeleton className="hidden h-5 w-16 md:block" />
                <Skeleton className="hidden h-5 w-20 md:block" />
                <Skeleton className="size-8 rounded-full" />
              </Flex>
            </RowWrapper>
          ))}
          {/* Group footer/summary row */}
          <RowWrapper className="pointer-events-none border-0 pt-4 pb-1.5">
            <Skeleton className="h-4 w-64" />
          </RowWrapper>
        </Box>
      ))}
    </Box>
  );
};

const KanbanLayoutSkeleton = () => {
  return (
    <Box>
      {/* Kanban header */}
      <Box className="sticky top-0 z-1 h-14 w-max px-6 backdrop-blur">
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
              <Skeleton className="shadow-shadow mb-3 h-28 shadow-sm" key={j} />
            ))}
          </Box>
        ))}
      </Box>
    </Box>
  );
};

const GanttLayoutSkeleton = () => {
  return (
    <div className="relative left-px h-[calc(100dvh-4rem)] overflow-x-auto overflow-y-hidden">
      <Flex className="min-h-full min-w-max">
        {/* Sidebar */}
        <Box className="border-border sticky left-0 z-20 w-136 shrink-0 border-r-[0.5px]">
          {/* Header */}
          <Flex
            align="center"
            className="border-border sticky top-0 z-10 h-16 border-b-[0.5px] px-6 py-2.5"
            justify="between"
          >
            <Skeleton className="h-8 w-16" />
            <Flex align="center" gap={2}>
              <Skeleton className="h-4 w-10" />
              <Skeleton className="h-8 w-20" />
            </Flex>
          </Flex>

          {/* Sidebar rows */}
          {Array.from({ length: 8 }).map((_, i) => (
            <Box className="border-border h-14 border-b-[0.5px]" key={i}>
              <Flex align="center" className="h-full px-6" justify="between">
                <Flex align="center" gap={3}>
                  <Skeleton className="h-4 w-4 rounded" />
                  <Skeleton className="h-4 w-4 rounded" />
                  <Skeleton className="h-4 w-40" />
                </Flex>
                <Flex align="center" gap={2}>
                  <Skeleton className="h-6 w-6 rounded-full" />
                  <Skeleton className="h-4 w-16" />
                </Flex>
              </Flex>
            </Box>
          ))}
        </Box>

        {/* Chart */}
        <Box className="flex-1" style={{ minWidth: "1344px" }}>
          {/* Timeline Header */}
          <Box className="border-border sticky top-0 z-10 h-16 border-b-[0.5px]">
            <Box className="h-8 w-full">
              {/* Month/Quarter row */}
              <Box className="border-border border-b-[0.5px]">
                <Flex>
                  {Array.from({ length: 3 }).map((_, i) => (
                    <Box
                      className="border-border border-r-[0.5px] px-2 py-1.5 text-left"
                      key={i}
                      style={{ width: "33.33%" }}
                    >
                      <Flex
                        align="center"
                        className="h-5 min-h-0"
                        justify="between"
                      >
                        <Skeleton className="h-4 w-8" />
                        <Skeleton className="h-4 w-6" />
                      </Flex>
                    </Box>
                  ))}
                </Flex>
              </Box>

              {/* Days/Periods row */}
              <Flex>
                {Array.from({ length: 21 }).map((_, i) => (
                  <Box
                    className="border-border h-[calc(2rem-1px)] min-w-16 flex-1 border-r-[0.5px] px-1 py-1 text-center"
                    key={i}
                    style={{ minWidth: "64px" }}
                  >
                    <Flex align="center" className="px-1" justify="between">
                      <Skeleton className="h-4 w-2" />
                      <Skeleton className="h-4 w-3" />
                    </Flex>
                  </Box>
                ))}
              </Flex>
            </Box>
          </Box>

          {/* Chart rows */}
          {Array.from({ length: 8 }).map((_, i) => (
            <Box className="relative h-14" key={i}>
              {/* Grid lines spanning full height */}
              <Flex className="absolute inset-0">
                {Array.from({ length: 21 }).map((_, j) => (
                  <Box
                    className="border-border min-w-16 flex-1 border-r-[0.5px]"
                    key={j}
                    style={{
                      minWidth: "64px",
                      height: i === 7 ? `calc(100dvh - 4rem - 4rem)` : "100%",
                    }}
                  />
                ))}
              </Flex>

              {/* Gantt bar */}
              <Box className="relative z-10 h-full px-2">
                <Skeleton
                  className="border-border absolute h-10 rounded-[0.6rem] border-[0.5px]"
                  style={{
                    left: `${Math.random() * 50 + 10}%`,
                    width: `${15 + Math.random() * 25}%`,
                    top: "6px",
                  }}
                />
              </Box>
            </Box>
          ))}
        </Box>
      </Flex>
    </div>
  );
};

export const BoardSkeleton = ({
  layout,
  className,
}: {
  layout: StoriesLayout;
  className?: string;
}) => {
  return (
    <BodyContainer
      className={cn(
        {
          "bg-surface-muted/50 overflow-x-auto": layout === "kanban",
          "bg-background overflow-auto": layout === "gantt",
        },
        className,
      )}
    >
      {layout === "kanban" && <KanbanLayoutSkeleton />}
      {layout === "gantt" && <GanttLayoutSkeleton />}
      {(layout === "list" || !layout) && <ListLayoutSkeleton />}
    </BodyContainer>
  );
};
