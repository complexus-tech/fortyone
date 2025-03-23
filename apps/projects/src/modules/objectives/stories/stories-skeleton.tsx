"use client";
import { Box } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { BoardSkeleton } from "@/components/ui/board-skeleton";

export const StoriesSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  return (
    <Box className="h-[calc(100vh-7.7rem)]">
      <BoardSkeleton layout={layout} />
    </Box>
  );
};
