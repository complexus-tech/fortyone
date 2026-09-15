"use client";
import { cn } from "lib";
import type { StoriesLayout } from "@/components/ui";
import { BoardSkeleton } from "@/components/ui/board-skeleton";

export const StoriesSkeleton = ({ layout }: { layout: StoriesLayout }) => {
  return (
    <BoardSkeleton
      className={cn({
        "h-(--app-page-content-height)": layout === "kanban",
      })}
      layout={layout}
    />
  );
};
