"use client";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";

export default function Loading() {
  const [layout] = useLocalStorage<StoriesLayout>(
    "teams:stories:layout",
    "kanban",
  );
  return <StoriesSkeleton layout={layout} />;
}
