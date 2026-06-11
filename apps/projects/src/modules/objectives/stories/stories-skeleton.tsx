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
    <Box className={cn("h-[calc(100vh-7.7rem)]", className)}>
      <BoardSkeleton layout={layout} />
    </Box>
  );
};
