"use client";

import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Skeleton } from "@/modules/profile/components/skeleton";

export default function Loading() {
  const [layout] = useLocalStorage<StoriesLayout>(
    "profile:stories:layout",
    "kanban",
  );
  return <Skeleton layout={layout} />;
}
