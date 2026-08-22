import { Box, Flex, Skeleton } from "ui";
import type { CalendarView } from "./calendar-view";

const DEFAULT_VISIBLE_START_HOUR = 8;
const DEFAULT_VISIBLE_END_HOUR = 21;
const HOUR_HEIGHT = 52;
const TIME_RAIL_WIDTH = 8;

export const CalendarGridSkeleton = ({
  view = "week",
}: {
  view?: CalendarView;
}) => {
  if (view === "month") {
    return (
      <Box className="grid min-h-0 flex-1 grid-cols-7 grid-rows-6 overflow-hidden">
        {Array.from({ length: 42 }).map((_, index) => (
          <Box
            className="border-border/60 border-r-[0.5px] border-b-[0.5px] p-4"
            key={index}
          >
            <Skeleton className="mb-4 h-9 w-9 rounded-full" />
            <Skeleton className="mb-2 h-6 w-4/5" />
            <Skeleton className="h-6 w-2/3" />
          </Box>
        ))}
      </Box>
    );
  }

  const dayCount = view === "day" ? 1 : 7;
  const dayColumn = dayCount === 1 ? "minmax(0, 1fr)" : "minmax(9.5rem, 1fr)";
  const gridTemplateColumns = `${TIME_RAIL_WIDTH}rem repeat(${dayCount}, ${dayColumn})`;

  return (
    <Box className="min-h-0 flex-1 overflow-hidden">
      <Box
        className="border-border/70 grid h-18 border-b-[0.5px]"
        style={{ gridTemplateColumns }}
      >
        <Box />
        {Array.from({ length: dayCount }).map((_, index) => (
          <Box
            className="border-border/60 flex items-center justify-center border-l-[0.5px] px-3 py-4"
            key={index}
          >
            <Skeleton className="h-12 w-12 rounded-full" />
          </Box>
        ))}
      </Box>
      <Box className="grid" style={{ gridTemplateColumns }}>
        <Box className="border-border/60 border-r-[0.5px]" />
        {Array.from({ length: dayCount }).map((_, dayIndex) => (
          <Box
            className="border-border/60 relative border-l-[0.5px]"
            key={dayIndex}
            style={{
              height: `${(DEFAULT_VISIBLE_END_HOUR - DEFAULT_VISIBLE_START_HOUR) * HOUR_HEIGHT}px`,
            }}
          >
            {Array.from({ length: 4 }).map((__, index) => (
              <Skeleton
                className="absolute right-3 left-3 h-14"
                key={index}
                style={{ top: `${(index * 2 + 1) * HOUR_HEIGHT}px` }}
              />
            ))}
          </Box>
        ))}
      </Box>
    </Box>
  );
};

const CalendarToolbarSkeleton = () => (
  <Flex
    align="center"
    className="border-border/70 h-16 shrink-0 gap-5 overflow-hidden border-b-[0.5px] px-5 py-3"
    justify="between"
  >
    <Flex align="center" className="shrink-0" gap={3}>
      <Flex align="center" gap={1}>
        <Skeleton className="h-8 w-8 rounded-lg" />
        <Skeleton className="h-8 w-8 rounded-lg" />
      </Flex>
      <Skeleton className="h-7 w-44" />
    </Flex>
    <Flex align="center" className="shrink-0" gap={2}>
      <Skeleton className="h-8 w-16" />
      <Skeleton className="h-8 w-20" />
      <Skeleton className="hidden h-8 w-36 md:block" />
    </Flex>
  </Flex>
);

export const CalendarContentSkeleton = ({
  view = "week",
}: {
  view?: CalendarView;
}) => (
  <Box
    aria-busy="true"
    aria-label="Loading calendar"
    className="bg-background flex h-[calc(100%-3.6rem)] min-h-0 flex-col overflow-hidden"
    role="status"
  >
    <CalendarToolbarSkeleton />
    <CalendarGridSkeleton view={view} />
  </Box>
);

export const CalendarRouteSkeleton = () => (
  <Box className="bg-background flex h-dvh flex-col overflow-hidden">
    <Flex
      align="center"
      className="border-border h-[3.6rem] w-full shrink-0 border-b-[0.5px] px-5 md:px-12"
      justify="between"
    >
      <Flex align="center" gap={2}>
        <Skeleton className="h-6 w-6 rounded-md md:hidden" />
        <Skeleton className="h-5 w-28" />
      </Flex>
      <Skeleton className="hidden h-8 w-36 md:block" />
    </Flex>
    <CalendarContentSkeleton />
  </Box>
);
