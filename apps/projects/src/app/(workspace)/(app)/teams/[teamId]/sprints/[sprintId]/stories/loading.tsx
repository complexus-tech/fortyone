"use client";
import { StoriesSkeleton } from "@/modules/sprints/stories/skeleton";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";

export default function Loading() {
  const [layout] = useLocalStorage<StoriesLayout>(
    "objective:sprints:layout",
    "kanban",
  );
  return <StoriesSkeleton layout={layout} />;
}
