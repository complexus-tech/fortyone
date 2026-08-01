import { cn } from "lib";
import { Box, Flex, Skeleton } from "ui";
import {
  GOAL_NODE_WIDTH,
  OBJECTIVE_NODE_WIDTH,
  PILLAR_NODE_WIDTH,
} from "./strategy-map-layout";

const SKELETON_CANVAS_WIDTH = 1400;

const cardClassName =
  "border-border-strong/45 bg-white/70 dark:border-border-strong/65 dark:bg-surface-elevated/45 absolute rounded-[14px] border-2 backdrop-blur";

const GoalSkeleton = () => (
  <Box
    className={cn(cardClassName, "px-7 py-6")}
    style={{
      height: 196,
      left: (SKELETON_CANVAS_WIDTH - GOAL_NODE_WIDTH) / 2,
      top: 72,
      width: GOAL_NODE_WIDTH,
    }}
  >
    <Skeleton className="mx-auto h-3 w-28 rounded" />
    <Skeleton className="mx-auto mt-4 h-6 w-64" />
    <Skeleton className="mx-auto mt-3 h-4 w-80" />
    <Skeleton className="mx-auto mt-2 h-4 w-56" />
    <Flex
      align="center"
      className="border-border mt-5 gap-0 border-t pt-4"
      justify="center"
    >
      {Array.from({ length: 3 }).map((_, index) => (
        <Box
          className={cn(
            "min-w-0 flex-1 text-center",
            index > 0 && "border-border border-l",
          )}
          key={index}
        >
          <Skeleton className="mx-auto h-5 w-8" />
          <Skeleton className="mx-auto mt-2 h-2.5 w-16 rounded" />
        </Box>
      ))}
    </Flex>
  </Box>
);

const pillarPositions = [70, 530, 990] as const;

const PillarSkeleton = ({ left }: { left: number }) => (
  <Box
    className={cn(cardClassName, "px-5 py-5")}
    style={{ height: 150, left, top: 340, width: PILLAR_NODE_WIDTH }}
  >
    <Skeleton className="h-3 w-28 rounded" />
    <Skeleton className="mt-3 h-5 w-52" />
    <Skeleton className="mt-2 h-4 w-64" />
    <Skeleton className="mt-2 h-4 w-44" />
    <Flex align="center" className="mt-3 gap-2">
      <Skeleton className="h-4 w-4 rounded-full" />
      <Skeleton className="h-3.5 w-20" />
    </Flex>
  </Box>
);

const objectivePositions = [70, 530, 990] as const;

const ObjectiveSkeleton = ({ left }: { left: number }) => (
  <Box
    className={cn(cardClassName, "px-4 py-4")}
    style={{ height: 222, left, top: 620, width: OBJECTIVE_NODE_WIDTH }}
  >
    <Flex align="start" className="gap-4" justify="between">
      <Box className="min-w-0 flex-1">
        <Skeleton className="h-5 w-full" />
        <Skeleton className="mt-2 h-5 w-3/4" />
      </Box>
      <Skeleton className="h-4 w-14 shrink-0" />
    </Flex>
    <Flex align="center" className="mt-4 gap-1.5">
      <Skeleton className="h-7 w-7 rounded-lg" />
      <Skeleton className="h-7 w-20 rounded-xl" />
      <Skeleton className="h-7 w-16 rounded-xl" />
      <Skeleton className="h-7 w-14 rounded-xl" />
    </Flex>
    <Flex align="center" className="mt-2 gap-1.5">
      <Skeleton className="h-7 w-24 rounded-xl" />
      <Skeleton className="h-7 w-24 rounded-xl" />
    </Flex>
    <Box className="border-border mt-4 border-t pt-3">
      <Flex align="center" justify="between">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-4 rounded" />
      </Flex>
      <Flex
        align="center"
        className="border-border/70 mt-3 gap-2.5 border-t pt-3"
      >
        <Skeleton className="h-3.5 w-3.5 rounded-full" />
        <Skeleton className="h-4 w-52" />
      </Flex>
    </Box>
  </Box>
);

const StrategySkeletonConnections = () => (
  <svg
    aria-hidden
    className="pointer-events-none absolute inset-0 h-full w-full"
    fill="none"
    viewBox={`0 0 ${SKELETON_CANVAS_WIDTH} 1000`}
  >
    <g
      opacity="0.6"
      stroke="var(--color-border-strong)"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2.25"
      vectorEffect="non-scaling-stroke"
    >
      <path d="M700 268 V304 H240 V340" />
      <path d="M700 268 V340" />
      <path d="M700 304 H1160 V340" />
      <path d="M240 490 V620" />
      <path d="M700 490 V620" />
      <path d="M1160 490 V620" />
    </g>
  </svg>
);

export const StrategyMapSkeleton = () => (
  <Box
    aria-label="Loading strategy map"
    className="bg-surface-muted/20 dark:bg-surface-elevated/35 relative h-full overflow-hidden"
    role="status"
  >
    <span className="sr-only">Loading strategy map</span>
    <Box
      className="bg-surface-muted/15 dark:bg-surface-elevated/25 absolute top-0 left-1/2 h-[1000px] -translate-x-1/2"
      style={{
        backgroundImage:
          "radial-gradient(var(--color-border-strong) 1.15px, transparent 1.15px)",
        backgroundSize: "22px 22px",
        width: SKELETON_CANVAS_WIDTH,
      }}
    >
      <StrategySkeletonConnections />
      <GoalSkeleton />
      {pillarPositions.map((left) => (
        <PillarSkeleton key={left} left={left} />
      ))}
      {objectivePositions.map((left) => (
        <ObjectiveSkeleton key={left} left={left} />
      ))}
    </Box>
  </Box>
);
