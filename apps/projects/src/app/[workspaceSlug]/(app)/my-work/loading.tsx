"use client";

import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { MyWorkSkeleton } from "@/modules/my-work/components/my-work-skeleton";

export default function Loading() {
  const [layout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "kanban",
  );
  return <MyWorkSkeleton layout={layout} />;
}
