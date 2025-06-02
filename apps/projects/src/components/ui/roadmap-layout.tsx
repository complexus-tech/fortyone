"use client";

import { Box, Button, Flex } from "ui";
import { cn } from "lib";
import { GanttIcon, ListIcon } from "icons";

export type RoadmapLayoutType = "list" | "gantt";

type RoadmapLayoutProps = {
  layout: RoadmapLayoutType;
  onLayoutChange: (layout: RoadmapLayoutType) => void;
  children: React.ReactNode;
  className?: string;
};

export const RoadmapLayout = ({
  layout,
  onLayoutChange,
  children,
  className,
}: RoadmapLayoutProps) => {
  return (
    <Box className={cn("h-full w-full", className)}>
      <Flex
        align="center"
        className="border-b-[0.5px] border-gray-200/60 bg-white px-6 py-3 dark:border-dark-100 dark:bg-dark"
        justify="between"
      >
        <Flex align="center" gap={2}>
          <Button
            className={cn(
              "h-8 w-8 px-0",
              layout === "list" && "bg-gray-100 dark:bg-dark-200",
            )}
            color="tertiary"
            onClick={() => {
              onLayoutChange("list");
            }}
          >
            <ListIcon className="h-4 w-4" />
          </Button>
          <Button
            className={cn(
              "h-8 w-8 px-0",
              layout === "gantt" && "bg-gray-100 dark:bg-dark-200",
            )}
            color="tertiary"
            onClick={() => {
              onLayoutChange("gantt");
            }}
          >
            <GanttIcon className="h-4 w-4" />
          </Button>
        </Flex>
      </Flex>
      <Box className="h-[calc(100%-4rem)] overflow-hidden">{children}</Box>
    </Box>
  );
};
