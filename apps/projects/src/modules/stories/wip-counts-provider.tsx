"use client";

import { useMemo } from "react";
import type { WorkflowCountsProviderProps } from "@/shared/story/board-property-slots";
import { useGroupedStories } from "./hooks/use-grouped-stories";

export const WorkflowCountsProvider = ({
  teamId,
  enabled,
  children,
}: WorkflowCountsProviderProps) => {
  const { data } = useGroupedStories(
    { groupBy: "status", ...(teamId ? { teamIds: [teamId] } : {}) },
    { enabled },
  );
  const counts = useMemo(
    () => new Map(data?.groups.map((group) => [group.key, group.totalCount])),
    [data],
  );
  return children(counts);
};
