"use client";

import { Box, Skeleton } from "ui";
import { useLocalStorage, useMediaQuery } from "@/hooks";
import { MyWorkSkeleton } from "@/modules/my-work/components/my-work-skeleton";
import type { MyWorkLayout } from "@/modules/my-work/types";

export default function Loading() {
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [layout] = useLocalStorage<MyWorkLayout>(
    "my-stories:stories:layout",
    "kanban",
  );
  if (layout === "calendar" && !isMobile) {
    return (
      <Box className="bg-surface h-[calc(100dvh-4rem)] px-6 py-5">
        <Skeleton className="mb-4 h-9 w-72" />
        <Skeleton className="h-[calc(100%-3.25rem)] w-full" />
      </Box>
    );
  }
  return <MyWorkSkeleton layout={layout === "calendar" ? "list" : layout} />;
}
