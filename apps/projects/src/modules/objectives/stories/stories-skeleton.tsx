"use client";
import { cn } from "lib";
import { Box } from "ui";
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
    <Box className={cn("h-full min-h-0", className)}>
      <BoardSkeleton className="h-full min-h-0" layout={layout} />
    </Box>
  );
};
