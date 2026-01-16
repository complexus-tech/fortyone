"use client";
import { useLocalStorage } from "@/hooks";
import { ObjectivePageSkeleton } from "@/modules/objectives/stories/objective-page-skeleton";
import type { StoriesLayout } from "@/components/ui";

export default function Loading() {
  const [layout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "kanban",
  );

  return <ObjectivePageSkeleton layout={layout} />;
}
