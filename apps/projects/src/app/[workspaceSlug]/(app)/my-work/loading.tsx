"use client";

import { useLocalStorage } from "@/hooks";
import { MyWorkSkeleton } from "@/modules/my-work/components/my-work-skeleton";
import { normalizeMyWorkLayout } from "@/modules/my-work/types";

export default function Loading() {
  const [layout] = useLocalStorage<string>(
    "my-stories:stories:layout",
    "kanban",
  );
  return <MyWorkSkeleton layout={normalizeMyWorkLayout(layout, "kanban")} />;
}
