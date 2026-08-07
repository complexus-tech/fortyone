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
      <Box className="flex h-[calc(100dvh-4rem)] flex-col overflow-hidden">
        <Skeleton className="h-18 w-full shrink-0 rounded-none" />
        <Skeleton className="min-h-0 w-full flex-1 rounded-none" />
      </Box>
    );
  }
  return <MyWorkSkeleton layout={layout === "calendar" ? "list" : layout} />;
}
