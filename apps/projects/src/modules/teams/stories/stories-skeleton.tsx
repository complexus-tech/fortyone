"use client";
import { cn } from "lib";
import { Box, Text } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { BoardSkeleton } from "@/components/ui/board-skeleton";

export const StoriesSkeleton = ({
  className,
  layout,
}: {
  className?: string;
  layout: StoriesLayout;
}) => {
  return (
    <>
      <Text className="sr-only" role="status">
        Loading tasks...
      </Text>
      <Box aria-hidden className={className} inert>
        <BoardSkeleton
          className={cn(
            {
              "h-(--app-page-content-height)":
                layout === "kanban" && !className,
            },
            className,
          )}
          layout={layout}
        />
      </Box>
    </>
  );
};
